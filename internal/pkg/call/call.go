package call

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

type BackedErr struct {
	Err      error
	ExitCode int
}

func (err *BackedErr) Error() string {
	return fmt.Sprintf("Error: %v, exit code: %d\n", err.Err, err.ExitCode)
}

type backend interface {
	runAsCmd(context.Context) ([]byte, error)
	getAsString() (string, error)
	cleanUp() error
}

func getBackend(callData *data.Call) (backend, error) {
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

func CallBackend(ctx context.Context, callData *data.Call, leaveTmpFile, printCommand bool) (string, error) {
	backendTimeoutContext := ctx
	if callData.Config.Timeout > -1 {
		backendTimeoutContext, _ = context.WithTimeout(ctx, time.Duration(callData.Config.Timeout)*time.Second)
	}

	backend, err := getBackend(callData)
	if err != nil {
		return "", errors.Wrapf(err, "Could not instantiate backend: %s", callData.Backend)
	}

	if printCommand {
		return backend.getAsString()
	}

	output, err := backend.runAsCmd(backendTimeoutContext)

	var removeTmpFileErr error
	if !leaveTmpFile || err != nil {
		if err := backend.cleanUp(); err != nil {
			removeTmpFileErr = errors.Wrap(err, "Could not remove file with [Body] contents")
		}
	}

	if backendTimeoutContext.Err() == context.DeadlineExceeded {
		// !! TODO !! Have string representation of the cmd in the error
		return "", utils.CascadeErrorMessage(
			errors.Errorf("Backend-call: %s timed out after %d seconds", callData.Backend, callData.Config.Timeout),
			removeTmpFileErr,
		)
	}

	if err != nil {
		return "", utils.CascadeErrorMessage(
			errors.Wrapf(err, "Error running: %s\n%s", callData.Backend, strings.TrimSpace(string(output))),
			removeTmpFileErr,
		)
	}

	if removeTmpFileErr != nil {
		return "", errors.Wrapf(removeTmpFileErr, "Error running: %s\n%s", callData.Backend, strings.TrimSpace(string(output)))
	}

	return string(output), nil
}
