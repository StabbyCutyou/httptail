package httptail

import (
	"io"
	"net/http"
)

type HttpRedirectLog struct {
	URL          string `json:"url"`
	RedirectType string `json:"redirect_type"`
	RedirectTo   string `string:"redirect_to"`
}

type HttpTailer struct {
	url       string
	redirects []HttpRedirectLog
	client    *http.Client
}

func NewHttpTailer(url string) (*HttpTailer, error) {
	t := &HttpTailer{
		url:       url,
		redirects: make([]HttpRedirectLog, 0),
	}
	// must be a more elegant way to do this but shrug
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
		// is the requests response status the one we want, or the one prior?
		// a problem for anothher day, this is to just prove the approach
		RedirectType: req.Response.Status,
	}
	t.redirects = append(t.redirects, l)
	return nil
}

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

func (t *HttpTailer) Results() []HttpRedirectLog {
	// this makes a copy because if you call Tail again, it resets the original slice
	// probably over thinking it but whatever
	l := make([]HttpRedirectLog, len(t.redirects))
	for i, r := range t.redirects {
		l[i] = r
	}
	return l
}
