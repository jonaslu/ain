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

type backendConstructor struct {
	BinaryName  string
	constructor func(*data.Call, string) backend
}

var ValidBackends = map[string]backendConstructor{
	"curl": {
		BinaryName:  "curl",
		constructor: newCurlBackend,
	},
	"httpie": {
		BinaryName:  "http",
		constructor: newHttpieBackend,
	},
	"wget": {
		BinaryName:  "wget",
		constructor: newWgetBackend,
	},
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
	requestedBackend := callData.Backend

	if backendConstructor, exists := ValidBackends[requestedBackend]; exists {
		return backendConstructor.constructor(callData, backendConstructor.BinaryName), nil
	}

	return nil, errors.Errorf("Unknown backend: %s", requestedBackend)
}

func ValidBackend(backendName string) bool {
	if _, exists := ValidBackends[backendName]; exists {
		return true
	}

	return false
}

func CallBackend(ctx context.Context, callData *data.Call, leaveTmpFile, printCommand bool) (string, error) {
	backendTimeoutContext := ctx
	if callData.Config.Timeout != data.TimeoutNotSet {
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
