package cli

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/frizinak/goscrape/output"
	"github.com/frizinak/goscrape/output/csv"
	"github.com/frizinak/goscrape/output/json"
	"github.com/frizinak/goscrape/output/tab"
)

type CLI struct {
	StdOut      output.Output
	StdErr      io.Writer
	Concurrency int
	Amount      int
	Timeout     time.Duration
	URLs        []*url.URL
	FieldList   []string
}

func Parse() (*CLI, error) {
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

	fields := flag.String(
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

	format := flag.String(
		"f",
		"tab",
		fmt.Sprintf(
			"Output format, one of [%s]",
			strings.Join(formatNames, ", "),
		),
	)

	concurrency := flag.Int("c", 8, "Concurrency")
	max := flag.Int("n", 0, "Maximum amount of urls to scrape")
	timeout := flag.Int("t", 5, "Http timeout in seconds")

	flag.Parse()
	baseRawUrls := flag.Args()
	if len(baseRawUrls) == 0 {
		return nil, fmt.Errorf("No urls specified")
	}
	baseUrls := make([]*url.URL, len(baseRawUrls))
	for i := range baseRawUrls {
		var err error
		baseUrls[i], err = url.Parse(baseRawUrls[i])
		if err != nil {
			return nil, fmt.Errorf(
				"Invalid url '%s': %s\n",
				baseRawUrls[i],
				err.Error(),
			)
		}
	}

	stdoutMaker, ok := formats[*format]
	if !ok {
		return nil, fmt.Errorf("Invalid format '%s'", *format)
	}

	if *concurrency < 1 {
		return nil, fmt.Errorf("Concurrency can not be lower than 1")
	}

	fieldList := strings.Split(*fields, ",")
	cli := &CLI{
		StdOut:      stdoutMaker(fieldList),
		StdErr:      os.Stderr,
		Concurrency: *concurrency,
		Amount:      *max,
		Timeout:     time.Duration(*timeout) * time.Second,
		URLs:        baseUrls,
		FieldList:   fieldList,
	}

	return cli, nil
}
