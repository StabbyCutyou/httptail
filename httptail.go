package httptail

import (
	"io"
	"net/http"
)

// HttpRedirect holds metadata about an individual redirect
type HttpRedirect struct {
	URL          string      `json:"url"`
	RedirectType string      `json:"redirect_type"`
	RedirectTo   string      `string:"redirect_to"`
	Headers      http.Header `json:"headers"`
	Trailers     http.Header `json:"trailers"`
}

// HttpRedirectLog holds structured results for each hop in the redirect chain
type HttpRedirectLog struct {
	OriginalURL   string         `json:"original_url"`
	FinalURL      string         `json:"final_url"`
	FinalHeaders  http.Header    `json:"final_headers"`
	FinalTrailers http.Header    `json:"final_trailers"`
	Redirects     []HttpRedirect `json:"redirects"`
}

func Tail(url string) (*HttpRedirectLog, error) {
	l := &HttpRedirectLog{
		OriginalURL: url,
		Redirects:   make([]HttpRedirect, 0),
	}
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			hr := HttpRedirect{
				URL:        via[len(via)-1].URL.String(),
				RedirectTo: req.URL.String(),
				// TODO - i was trying to get the status from the last request in
				// via, but the response body was gone by this time
				// is the requests response status the one we want, or the one in via?
				// a problem for another day, this is to just prove the approach
				RedirectType: req.Response.Status,
				Headers:      req.Response.Header,
				Trailers:     req.Response.Trailer,
			}
			l.Redirects = append(l.Redirects, hr)
			return nil
		},
	}
	// head calls only
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	// proper hygiene
	defer resp.Body.Close()
	l.FinalURL = resp.Request.URL.String()
	io.Copy(io.Discard, resp.Body)
	return l, nil
}
