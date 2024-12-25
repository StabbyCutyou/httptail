package httptail

import (
	"io"
	"net/http"
)

// HttpRedirectLog holds structured results for each hop in the redirect chain
type HttpRedirectLog struct {
	URL          string `json:"url"`
	RedirectType string `json:"redirect_type"`
	RedirectTo   string `string:"redirect_to"`
}

// HttpTailer holds the metadata necessary to tail http redirects
type HttpTailer struct {
	url       string
	redirects []HttpRedirectLog
	client    *http.Client
}

// NewHttpTailer returns an HttpTailer ready to use for the url provided
func NewHttpTailer(url string) (*HttpTailer, error) {
	t := &HttpTailer{
		url:       url,
		redirects: make([]HttpRedirectLog, 0),
	}
	t.client = &http.Client{
		CheckRedirect: t.checkRedirect,
	}
	return t, nil
}

func (t *HttpTailer) checkRedirect(req *http.Request, via []*http.Request) error {
	l := HttpRedirectLog{
		URL:        via[len(via)-1].URL.String(),
		RedirectTo: req.URL.String(),
		// TODO - i was trying to get the status from the last request in
		// via, but the response body was gone by this time
		// is the requests response status the one we want, or the one in via?
		// a problem for another day, this is to just prove the approach
		RedirectType: req.Response.Status,
	}
	t.redirects = append(t.redirects, l)
	return nil
}

// Tail will begin the chain of HTTP calls using HEAD to follow redirects
// and populate an internal cache of hops made, for later inspection
func (t *HttpTailer) Tail() error {
	// clear out the old redirects first
	t.redirects = make([]HttpRedirectLog, 0)
	// head calls only
	r, err := http.NewRequest("HEAD", t.url, nil)
	if err != nil {
		return err
	}
	resp, err := t.client.Do(r)
	if err != nil {
		return err
	}
	// proper hygiene
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

// Results will return a copy of the internal results, as re-running Tail
// will reset them, and you may want to re-use the HttpTailer
func (t *HttpTailer) Results() []HttpRedirectLog {
	l := make([]HttpRedirectLog, len(t.redirects))
	for i, r := range t.redirects {
		l[i] = r
	}
	return l
}
