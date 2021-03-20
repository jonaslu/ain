package call

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
)

type Config struct {
	Timeout int32
}

type Data struct {
	Host    *url.URL
	Body    []string
	Method  string
	Headers []string

	Backend        string
	BackendOptions []string

	Config Config
}

type backend interface {
	runAsCmd(context.Context) ([]byte, error)
	// getAsString() string
	// cleanUp()
}

func (callData *Data) getBackend() (backend, error) {
	switch callData.Backend {
	case "httpie":
		return newHttpieBackend(callData)
	case "curl":
		return newCurlBackend(callData)
	}

	return nil, errors.Errorf("Unknown backend: %s", callData.Backend)
}

func ValidBackend(backendName string) bool {
	switch backendName {
	case "httpie":
		return true
	case "curl":
		return true
	}

	return false
}
