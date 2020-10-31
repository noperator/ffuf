package ffuf

import (
	"net/http"
	"net/url"
)

// Response struct holds the meaningful data returned from request and is meant for passing to filters
type Response struct {
	StatusCode    int64
	Headers       map[string][]string
	Data          []byte
	ContentLength int64
	ContentWords  int64
	ContentLines  int64
	Cancelled     bool
	Request       *Request
	Raw           string
	ResultFile    string
	RedirectChain []string
}

// GetRedirectLocation returns the redirect location for a 3xx redirect HTTP response
func (resp *Response) GetRedirectLocation(absolute bool) string {

	redirectLocation := ""
	if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
		if loc, ok := resp.Headers["Location"]; ok {
			if len(loc) > 0 {
				redirectLocation = loc[0]
			}
		}
	}

	if absolute {
		redirectUrl, err := url.Parse(redirectLocation)
		if err != nil {
			return redirectLocation
		}
		baseUrl, err := url.Parse(resp.Request.Url)
		if err != nil {
			return redirectLocation
		}
		redirectLocation = baseUrl.ResolveReference(redirectUrl).String()
	}

	return redirectLocation
}

func NewResponse(httpresp *http.Response, req *Request) Response {
	// Building a redirect chain by iterating over each Request's Response
	// until there are no more Responses left. Request.Response is the redirect
	// response which caused this request to be created. This field is
	// only populated during client redirects.
	// Source: https://golang.org/pkg/net/http/#Request
	redirectChain := []string{httpresp.Request.URL.String()}
	redirectReq := httpresp.Request.Response
	for redirectReq != nil {
		redirectChain = append([]string{redirectReq.Request.URL.String()}, redirectChain...)
		redirectReq = redirectReq.Request.Response
	}

	var resp Response
	resp.Request = req
	resp.StatusCode = int64(httpresp.StatusCode)
	resp.Headers = httpresp.Header
	resp.Cancelled = false
	resp.Raw = ""
	resp.ResultFile = ""
	resp.RedirectChain = redirectChain[1:]  // Remove the original URL.
	return resp
}
