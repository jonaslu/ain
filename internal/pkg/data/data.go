package data

import (
	"net/url"
)

const TimeoutNotSet = -1

type Config struct {
	Timeout int32
}

type Parse struct {
	Host    []string
	Body    []string
	Method  string
	Headers []string

	Backend        string
	BackendOptions [][]string

	Config Config
}

type Call struct {
	Host    *url.URL
	Body    []string
	Method  string
	Headers []string

	Backend        string
	BackendOptions [][]string

	Config Config
}
