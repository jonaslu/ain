package call

import (
	"context"
	"fmt"
	"strings"

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
	constructor func(*data.BackendInput, string) backend
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
	getAsString() string
}

func getBackend(backendInput *data.BackendInput) (backend, error) {
	requestedBackend := backendInput.Backend

	if backendConstructor, exists := ValidBackends[requestedBackend]; exists {
		return backendConstructor.constructor(backendInput, backendConstructor.BinaryName), nil
	}

	return nil, errors.Errorf("Unknown backend: %s", requestedBackend)
}

func ValidBackend(backendName string) bool {
	if _, exists := ValidBackends[backendName]; exists {
		return true
	}

	return false
}

func CallBackend(ctx context.Context, backendInput *data.BackendInput) (string, error) {
	backend, err := getBackend(backendInput)
	if err != nil {
		return "", errors.Wrapf(err, "Could not instantiate backend: %s", backendInput.Backend)
	}

	if err := backendInput.CreateBodyTempFile(); err != nil {
		return "", err
	}

	if backendInput.PrintCommand {
		return backend.getAsString(), nil
	}

	output, err := backend.runAsCmd(ctx)
	removeTmpFileErr := backendInput.RemoveBodyTempFile(err != nil)

	if ctx.Err() == context.Canceled {
		return "", removeTmpFileErr
	}

	if ctx.Err() == context.DeadlineExceeded {
		return "", utils.CascadeErrorMessage(
			errors.Errorf("Backend-call: %s timed out after %d seconds", backendInput.Backend, ctx.Value(data.TimeoutContextValueKey{})),
			removeTmpFileErr,
		)
	}

	if err != nil {
		return "", utils.CascadeErrorMessage(
			errors.Wrapf(err, "Error running: %s\n%s", backendInput.Backend, strings.TrimSpace(string(output))),
			removeTmpFileErr,
		)
	}

	if removeTmpFileErr != nil {
		return "", errors.Wrapf(removeTmpFileErr, "Error running: %s\n%s", backendInput.Backend, strings.TrimSpace(string(output)))
	}

	return string(output), nil
}
