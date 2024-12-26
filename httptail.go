package httptail

import (
	"io"
	"net/http"
	"net/http/httptrace"
)

// HttpRedirect holds metadata about an individual redirect
type HttpRedirect struct {
	URL          string      `json:"url"`
	IP           string      `json:"ip"`
	RedirectType string      `json:"redirect_type"`
	RedirectTo   string      `string:"redirect_to"`
	Headers      http.Header `json:"headers"`
	Trailers     http.Header `json:"trailers"`
}

// HttpRedirectLog holds structured results for each hop in the redirect chain
type HttpRedirectLog struct {
	OriginalURL string         `json:"original_url"`
	FinalURL    string         `json:"final_url"`
	Redirects   []HttpRedirect `json:"redirects"`
}

func Tail(url string) (*HttpRedirectLog, error) {
	l := &HttpRedirectLog{
		OriginalURL: url,
		Redirects:   make([]HttpRedirect, 0),
	}
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// By now, the trace resolver has already set up the redirect
			// etnry we need with the IP, so this captures the rest of that data
			hr := l.Redirects[len(l.Redirects)-1]
			hr.URL = via[len(via)-1].URL.String()
			hr.RedirectTo = req.URL.String()
			// TODO - i was trying to get the status from the last request in
			// via, but the response body was gone by this time
			// is the requests response status the one we want, or the one in via?
			// a problem for another day, this is to just prove the approach
			hr.RedirectType = req.Response.Status
			hr.Headers = req.Response.Header
			hr.Trailers = req.Response.Trailer
			l.Redirects[len(l.Redirects)-1] = hr
			return nil
		},
	}
	// head calls only
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	t := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			// TODO what else can we record here?
			hr := HttpRedirect{
				IP: connInfo.Conn.RemoteAddr().String(),
			}
			l.Redirects = append(l.Redirects, hr)
		},
	}
	r = r.WithContext(httptrace.WithClientTrace(r.Context(), t))
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	// proper hygiene
	defer resp.Body.Close()
	// Record some final metadata about the end request
	f := l.Redirects[len(l.Redirects)-1]
	f.URL = resp.Request.URL.String()
	f.Headers = resp.Header
	f.Trailers = resp.Trailer
	l.Redirects[len(l.Redirects)-1] = f
	// Very easy top level reporting
	l.FinalURL = f.URL
	io.Copy(io.Discard, resp.Body)
	return l, nil
}
