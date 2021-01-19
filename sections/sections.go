package sections

import "net/url"

type TemplateSections struct {
	Host       *url.URL
	URL        string
	Body       string
	Parameters []string
	Method     string
	Headers    []string
}
