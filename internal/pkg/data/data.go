package data

import (
	"net/url"
)

const TimeoutNotSet = -1

type Config struct {
	Timeout    int32
	QueryDelim *string
}

func NewConfig() Config {
	return Config{Timeout: TimeoutNotSet}
}

type BackendInput struct {
	Host    *url.URL
	Body    []string
	Method  string
	Headers []string

	Backend        string
	BackendOptions [][]string

	PrintCommand  bool
	LeaveTempFile bool

	TempFileName string
}

type TimeoutContextValueKey struct{}
