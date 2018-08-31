package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/frizinak/goscrape/cli"
	"github.com/frizinak/goscrape/fetcher"
	"github.com/frizinak/goscrape/output"
)

type task struct {
	from *url.URL
	to   *url.URL
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

	f := fetcher.New(cli.Timeout)
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

	for _, u := range cli.URLs {
		work <- &task{nil, u}
	}

	canceled := false
	fields := make([]fmt.Stringer, len(cli.FieldList))
	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.Err != nil {
			fmt.Fprintln(cli.StdErr, r.URL, r.Err)
			continue
		}

		if r.Status == 0 || r.Status >= 400 {
			for i := range cli.FieldList {
				fields[i] = r.GetString(cli.FieldList[i], output.NewString("-"))
			}

			cli.StdOut.Write(fields)
		}

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
			if cli.Amount > 0 && len(fetched) >= cli.Amount {
				canceled = true
				close(work)
				break
			}
		}
	}
}
