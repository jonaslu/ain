package call

import (
	"context"
	"time"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/pkg/errors"
)

type backend interface {
	runAsCmd(context.Context) ([]byte, error)
	// getAsString() string
	// cleanUp()
}

func getBackend(callData *data.Parse) (backend, error) {
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

func CallBackend(ctx context.Context, callData *data.Parse) (string, error) {
	backendTimeoutContext := ctx
	if callData.Config.Timeout > -1 {
		backendTimeoutContext, _ = context.WithTimeout(ctx, time.Duration(callData.Config.Timeout)*time.Second)
	}

	backend, err := getBackend(callData)
	if err != nil {
		return "", errors.Wrapf(err, "Could not instantiate backend: %s", callData.Backend)
	}

	output, err := backend.runAsCmd(backendTimeoutContext)
	if backendTimeoutContext.Err() == context.DeadlineExceeded {
		// !! TODO !! Have string representation of the cmd in the error
		return "", errors.Wrapf(err, "Backend-call: %s timed out after %d seconds", callData.Backend, callData.Config.Timeout)
	}

	if err != nil {
		errors.Wrapf(err, "Error running: %s", callData.Backend)
	}

	return string(output), nil
}
