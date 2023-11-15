package call

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

type Output struct {
	StdOut []byte
	StdErr []byte
}

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
	generateCmd(context.Context) (*exec.Cmd, error)
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

func CallBackend(ctx context.Context, callData *data.Call, leaveTmpFile, printCommand bool) (Output, error) {
	backendTimeoutContext := ctx
	if callData.Config.Timeout != data.TimeoutNotSet {
		backendTimeoutContext, _ = context.WithTimeout(ctx, time.Duration(callData.Config.Timeout)*time.Second)
	}

	backend, err := getBackend(callData)
	if err != nil {
		return Output{}, errors.Wrapf(err, "Could not instantiate backend: %s", callData.Backend)
	}

	if printCommand {
		if command, err := backend.getAsString(); err != nil {
			return Output{StdErr: []byte(command)}, err
		} else {
			return Output{StdOut: []byte(command)}, nil
		}
	}

	cmd, err := backend.generateCmd(backendTimeoutContext)
        if err != nil {
            return Output{}, errors.Wrapf(err, "Could not generate valid command: %s", callData.Backend)
        }

        output, err := runCmd(cmd);

	var removeTmpFileErr error
	if !leaveTmpFile || err != nil {
		if err := backend.cleanUp(); err != nil {
			removeTmpFileErr = errors.Wrap(err, "Could not remove file with [Body] contents")
		}
	}

	if backendTimeoutContext.Err() == context.DeadlineExceeded {
		return output, utils.CascadeErrorMessage(
			errors.Errorf("Backend-call: %s timed out after %d seconds", callData.Backend, callData.Config.Timeout),
			removeTmpFileErr,
		)
	}

	if err != nil {
		return output, utils.CascadeErrorMessage(
			errors.Wrapf(err, "Error running: %s\n%s", callData.Backend, strings.TrimSpace(string(output.StdOut))),
			removeTmpFileErr,
		)
	}

	if removeTmpFileErr != nil {
		return output, errors.Wrapf(removeTmpFileErr, "Error running: %s\n%s", callData.Backend, strings.TrimSpace(string(output.StdOut)))
	}

	return output, nil
}

func runCmd(cmd *exec.Cmd) (Output, error) {
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return Output{}, err
	}
	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return Output{}, err
	}
	cmd.Start()
	if err != nil {
		return Output{}, err
	}
	stdOut, err := io.ReadAll(stdOutPipe)
	if err != nil {
		return Output{StdOut: stdOut}, &BackedErr{
			Err:      err,
			ExitCode: cmd.ProcessState.ExitCode(),
		}
	}
	stdErr, err := io.ReadAll(stdErrPipe)
	if err != nil {
		return Output{StdOut: stdOut, StdErr: stdErr}, &BackedErr{
			Err:      err,
			ExitCode: cmd.ProcessState.ExitCode(),
		}
	}
	err = cmd.Wait()
	if err != nil {
		return Output{StdOut: stdOut, StdErr: stdErr}, &BackedErr{
			Err:      err,
			ExitCode: cmd.ProcessState.ExitCode(),
		}
	}
	return Output{StdOut: stdOut, StdErr: stdErr}, nil
}

