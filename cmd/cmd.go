package cmd

import (
	"flag"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/frizinak/goscrape/fetcher"
	"github.com/frizinak/goscrape/output"
	"github.com/frizinak/goscrape/output/csv"
	"github.com/frizinak/goscrape/output/json"
	"github.com/frizinak/goscrape/output/tab"
)

type stats struct {
	errs        int
	success     int
	fastest     time.Duration
	slowest     time.Duration
	mean        time.Duration
	average     time.Duration
	statusCodes map[int]int
}

type task struct {
	from *url.URL
	to   *url.URL
}

type statusCode struct {
	code   int
	amount int
}

type statusCodes []statusCode

func (s statusCodes) Len() int           { return len(s) }
func (s statusCodes) Less(i, j int) bool { return s[i].code < s[j].code }
func (s statusCodes) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func calcStats(s *stats, timings []time.Duration) {
	if len(timings) == 0 {
		return
	}

	var avg float64 = 0
	for _, t := range timings {
		avg = (avg*float64(s.success) + float64(t)) / float64(s.success+1)
		s.success++
		if t < s.fastest {
			s.fastest = t
		}
		if t > s.slowest {
			s.slowest = t
		}
	}

	s.average = time.Duration(avg)

	var middle float64 = float64(len(timings)) / 2
	middleF := int(middle)

	if float64(middleF) == middle || len(timings) <= middleF+1 {
		s.mean = timings[middleF]
		return
	}

	s.mean = (timings[middleF] + timings[middleF+1]) / 2
}

func handleWork(
	f *fetcher.Fetcher,
	work <-chan *task,
	results chan<- *fetcher.Result,
	timeout time.Duration,
) {
	var rwg sync.WaitGroup
	defer rwg.Wait()

	for {
		select {
		case u, ok := <-work:
			if !ok {
				return
			}

			r := f.Fetch(u.to, u.from)
			rwg.Add(1)
			go func() { results <- r; rwg.Done() }()
		case <-time.After(timeout):
			return
		}
	}
}

func handleResults(
	fieldList []string,
	stderr io.Writer,
	stdout output.Output,
	results <-chan *fetcher.Result,
	work chan<- *task,
	fetched map[string]bool,
	max *int,
) (s *stats) {
	s = &stats{
		fastest:     time.Duration(math.MaxInt64),
		statusCodes: make(map[int]int),
	}

	timings := make([]time.Duration, 0)

	canceled := false
	f := make([]fmt.Stringer, len(fieldList))
	for r := range results {
		if r.Err != nil {
			s.errs++
			fmt.Fprintln(stderr, r.URL, r.Err)
			continue
		}

		timings = append(timings, r.Duration)
		s.statusCodes[r.Status]++

		for i := range fieldList {
			f[i] = r.GetString(fieldList[i], output.NewString("-"))
		}
		stdout.Write(f)

		if canceled {
			continue
		}

		for _, u := range r.Urls {
			str := u.String()
			if _, ok := fetched[str]; ok {
				continue
			}

			fetched[str] = true
			work <- &task{r.URL, u}
			if *max > 0 && len(fetched) >= *max {
				canceled = true
				close(work)
				break
			}
		}
	}

	calcStats(s, timings)
	return s
}

func Cmd(flags *flag.FlagSet, args []string, stderr io.Writer) (err error) {
	formats := map[string]func(fields []string) output.Output{
		"csv": func(fields []string) output.Output {
			return csv.New(os.Stdout, fields)
		},
		"tab": func(fields []string) output.Output {
			return tab.New(os.Stdout)
		},
		"json": func(fields []string) output.Output {
			return json.New(os.Stdout, fields)
		},
	}

	formatNames := make([]string, 0, len(formats))
	for i := range formats {
		formatNames = append(formatNames, i)
	}

	fields := flags.String(
		"o",
		"status,duration,path,query",
		`Comma separated list of fields.
		Available fields:
			url:        the request url
			path:       the request path
			query:      the request query params
			nurls:      amount of scrapable urls on the page
			origin:     the origin url
			originpath: the origin path
			status:     the http status code
			head:       the amount of time it took until headers were received
			duration:   the total amount of time it took until we received the entire response
			header.*:   replace * with the header to include in the output
			meta.*:     replace * with the meta property to include in the output
			query.*:    replace * with the query param to include in the output
			`,
	)

	format := flags.String(
		"f",
		"tab",
		fmt.Sprintf(
			"Output format, one of [%s]",
			strings.Join(formatNames, ", "),
		),
	)

	concurrency := flags.Int("c", 8, "Concurrency")
	max := flags.Int("n", 0, "Maximum amount of urls to scrape")
	timeout := flags.Int("t", 5, "Http timeout in seconds")

	flags.Parse(args)
	baseRawUrls := flags.Args()
	if len(baseRawUrls) == 0 {
		return fmt.Errorf("No urls specified")
	}

	baseUrls := make([]*url.URL, len(baseRawUrls))
	fieldList := strings.Split(*fields, ",")

	stdoutMaker, ok := formats[*format]
	if !ok {
		return fmt.Errorf("Invalid format '%s'\n", *format)
	}

	stdout := stdoutMaker(fieldList)

	if *concurrency < 1 {
		return fmt.Errorf("Concurrency can not be lower than 1")
	}

	fetched := make(map[string]bool)
	for i := range baseRawUrls {
		var err error
		baseUrls[i], err = url.Parse(baseRawUrls[i])
		if err != nil {
			return fmt.Errorf(
				"Invalid url '%s': %s\n",
				baseRawUrls[i],
				err.Error(),
			)
		}

		fetched[baseUrls[i].String()] = true
	}

	to := time.Duration(*timeout) * time.Second
	f := fetcher.New(to)

	workers := *concurrency
	work := make(chan *task, workers)
	results := make(chan *fetcher.Result, 100*workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			handleWork(f, work, results, to)
			wg.Done()
		}()
	}

	trap(max, stderr)

	for _, u := range baseUrls {
		work <- &task{nil, u}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	stats := handleResults(fieldList, stderr, stdout, results, work, fetched, max)
	fmt.Fprintf(
		stderr,
		`
Success:  %d
Errors:   %d

Fastest:  %s
Slowest:  %s

Mean:     %s
Average:  %s

StatusCodes:
`,
		stats.success,
		stats.errs,
		stats.fastest,
		stats.slowest,
		stats.mean,
		stats.average,
	)

	codes := make(statusCodes, 0, len(stats.statusCodes))
	for i := range stats.statusCodes {
		codes = append(codes, statusCode{i, stats.statusCodes[i]})
	}
	sort.Sort(codes)

	for _, c := range codes {
		fmt.Fprintf(stderr, "\t%03d: %-5d\n", c.code, c.amount)
	}

	return nil
}
