package data

import (
	"net/url"
)

type Config struct {
	Timeout int32
}

type Parse struct {
	Host    *url.URL
	Body    []string
	Method  string
	Headers []string

	Backend        string
	BackendOptions []string

	Config Config
}
