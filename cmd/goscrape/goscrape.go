package main

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	"github.com/frizinak/goscrape/cli"
	"github.com/frizinak/goscrape/fetcher"
	"github.com/frizinak/goscrape/output"
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
	cli *cli.CLI,
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
	f := make([]fmt.Stringer, len(cli.FieldList))
	for r := range results {
		if r.Err != nil {
			s.errs++
			fmt.Fprintln(cli.StdErr, r.URL, r.Err)
			continue
		}

		timings = append(timings, r.Duration)
		s.statusCodes[r.Status]++

		for i := range cli.FieldList {
			f[i] = r.GetString(cli.FieldList[i], output.NewString("-"))
		}
		cli.StdOut.Write(f)

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

func main() {
	cli, err := cli.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fetched := make(map[string]bool)
	for i := range cli.URLs {
		fetched[cli.URLs[i].String()] = true
	}

	f := fetcher.New(cli.Timeout, "")
	workers := cli.Concurrency
	work := make(chan *task, workers)
	results := make(chan *fetcher.Result, 100*workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			handleWork(f, work, results, cli.Timeout)
			wg.Done()
		}()
	}

	max := &cli.Amount
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		fmt.Fprintln(cli.StdErr, "quitting...")
		*max = 1
	}()

	for _, u := range cli.URLs {
		work <- &task{nil, u}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	stats := handleResults(
		cli,
		results,
		work,
		fetched,
		max,
	)
	fmt.Fprintf(
		cli.StdErr,
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
		fmt.Fprintf(cli.StdErr, "\t%03d: %-5d\n", c.code, c.amount)
	}
}
