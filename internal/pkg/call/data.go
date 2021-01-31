package call

import "net/url"

type Data struct {
	Host       *url.URL
	URL        string
	Body       string
	Parameters []string
	Method     string
	Headers    []string
}
