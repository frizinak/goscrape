package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/frizinak/goscrape/output"
)

var specialHrefRE *regexp.Regexp

func init() {
	specialHrefRE = regexp.MustCompile("^[a-zA-Z]+:")
}

type Result struct {
	Err      error
	URL      *url.URL
	Head     time.Duration
	Duration time.Duration
	Urls     []*url.URL
	Status   int
	Headers  http.Header
	Meta     map[string]string
}

func (r *Result) GetString(key string, fallback fmt.Stringer) fmt.Stringer {
	if strings.HasPrefix(key, "header.") {
		if h := r.Headers.Get(strings.SplitN(key, ".", 2)[1]); h != "" {
			return output.NewString(h)
		}

		return fallback
	}

	if strings.HasPrefix(key, "meta.") {
		if h, ok := r.Meta[strings.SplitN(key, ".", 2)[1]]; ok {
			return output.NewString(h)
		}

		return fallback
	}

	if strings.HasPrefix(key, "query.") {
		if k := r.URL.Query().Get(strings.SplitN(key, ".", 2)[1]); k != "" {
			return output.NewString(k)
		}

		return fallback
	}

	switch key {
	case "status":
		return output.NewInt(r.Status)
	case "head":
		return r.Head
	case "duration":
		return r.Duration
	case "url":
		return r.URL
	case "path":
		return output.NewString(r.URL.Path)
	case "query":
		if q := r.URL.Query().Encode(); q != "" {
			return output.NewString(q)
		}
	case "nurls":
		return output.NewInt(len(r.Urls))
	}

	return fallback
}

type Fetcher struct {
	client http.Client
}

func New(timeout time.Duration) *Fetcher {
	return &Fetcher{
		http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (f *Fetcher) Fetch(u *url.URL) *Result {
	r := &Result{URL: u}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		r.Err = err
		return r
	}

	start := time.Now()
	res, err := f.client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	if err != nil {
		r.Err = err
		return r
	}

	r.Head = time.Now().Sub(start)
	r.Headers = res.Header
	r.Status = res.StatusCode
	meta, urls, err := extract(u, res.Body)
	if err != nil {
		r.Err = err
		return r
	}

	r.Urls = urls
	r.Meta = meta
	r.Duration = time.Now().Sub(start)

	return r
}

func extract(
	baseURL *url.URL,
	body io.Reader,
) (map[string]string, []*url.URL, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, nil, err
	}

	if strings.HasSuffix(baseURL.Path, ".xml") {
		locs := doc.Find("urlset>url>loc")
		if locs.Length() != 0 {
			urls := extractSitemap(baseURL, locs)
			return map[string]string{}, urls, nil
		}
	}

	s := doc.Find("a[href]")
	m := doc.Find("head meta")
	urls := make([]*url.URL, 0, s.Length())
	meta := make(map[string]string, m.Length())

	s.Each(func(i int, s *goquery.Selection) {
		u, exists := s.Attr("href")
		if !exists {
			return
		}

		if p := normalize(baseURL, u); p != nil {
			urls = append(urls, p)
		}
	})

	m.Each(func(i int, s *goquery.Selection) {
		key, exists := s.Attr("property")
		if !exists {
			key, exists = s.Attr("name")
			if !exists {
				return
			}
		}

		val, _ := s.Attr("content")
		meta[key] = val
	})

	return meta, urls, nil
}

func extractSitemap(baseURL *url.URL, s *goquery.Selection) []*url.URL {
	urls := make([]*url.URL, 0, s.Length())
	s.Each(func(i int, s *goquery.Selection) {
		u := s.Text()
		if p := normalize(baseURL, u); p != nil {
			urls = append(urls, p)
		}
	})

	return urls
}

func normalize(baseURL *url.URL, u string) *url.URL {

	switch {
	case u == "" || strings.HasPrefix(u, "#"):
		return nil
	case strings.HasPrefix(u, "http:") || strings.HasPrefix(u, "https:"):
		break
	case specialHrefRE.MatchString(u):
		return nil
	case !strings.HasPrefix(u, "/"):
		u = baseURL.String() + "/" + u
	case strings.HasPrefix(u, "//"):
		u = baseURL.Scheme + ":" + u
	}

	p, err := url.Parse(u)
	if err != nil {
		return nil
	}

	if p.Host == "" {
		p.Host = baseURL.Host
	}

	if p.Host != baseURL.Host {
		return nil
	}

	p.Fragment = ""
	p.Scheme = baseURL.Scheme
	if baseURL.User != nil {
		p.User = baseURL.User
	}

	return p
}
