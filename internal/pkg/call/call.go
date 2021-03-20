package call

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

func CallBackend(ctx context.Context, callData *Data) (string, error) {
	backendTimeoutContext := ctx
	if callData.Config.Timeout > -1 {
		backendTimeoutContext, _ = context.WithTimeout(ctx, time.Duration(callData.Config.Timeout)*time.Second)
	}

	backend, err := callData.getBackend()
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
