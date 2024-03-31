package call

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/pkg/errors"
)

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

type backend interface {
	getAsCmd(context.Context) *exec.Cmd
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

type Call struct {
	backendInput        *data.BackendInput
	backend             backend
	forceRemoveTempFile bool
}

func Setup(backendInput *data.BackendInput) (*Call, error) {
	call := Call{backendInput: backendInput}

	backend, err := getBackend(backendInput)
	if err != nil {
		return nil, err
	}

	call.backend = backend

	if err := backendInput.CreateBodyTempFile(); err != nil {
		return nil, err
	}

	return &call, nil
}

func (c *Call) CallAsString() string {
	return c.backend.getAsString()
}

func (c *Call) CallAsCmd(ctx context.Context) (*data.BackendOutput, error) {
	backendCmd := c.backend.getAsCmd(ctx)

	var stdout, stderr bytes.Buffer
	backendCmd.Stdout = &stdout
	backendCmd.Stderr = &stderr

	err := backendCmd.Run()

	c.forceRemoveTempFile = err != nil

	backendOutput := &data.BackendOutput{
		Stderr:   stderr.String(),
		Stdout:   stdout.String(),
		ExitCode: backendCmd.ProcessState.ExitCode(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		err = errors.Errorf("Backend-call: %s timed out after %d seconds",
			c.backendInput.Backend,
			ctx.Value(data.TimeoutContextValueKey{}))

		return backendOutput, err
	}

	if err != nil {
		return backendOutput, errors.Wrapf(err, "Error running: %s", c.backendInput.Backend)
	}

	return backendOutput, nil
}

func (c *Call) Teardown() error {
	return c.backendInput.RemoveBodyTempFile(c.forceRemoveTempFile)
}
