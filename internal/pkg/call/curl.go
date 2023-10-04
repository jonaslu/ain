package call

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

type curl struct {
	callData    *data.Call
	tmpFileName string
	binaryName  string
}

func newCurlBackend(callData *data.Call, binaryName string) backend {
	return &curl{
		callData:   callData,
		binaryName: binaryName,
	}
}

func (curl *curl) getHeaderArguments(escape bool) [][]string {
	args := [][]string{}
	for _, header := range curl.callData.Headers {
		headerVal := header
		if escape {
			headerVal = utils.EscapeForShell(header)
		}

		args = append(args, []string{"-H", headerVal})
	}

	return args
}

func (curl *curl) getMethodArgument(escape bool) []string {
	if curl.callData.Method != "" {
		methodCapitalized := strings.ToUpper(curl.callData.Method)
		if escape {
			methodCapitalized = utils.EscapeForShell(methodCapitalized)
		}

		return []string{"-X", methodCapitalized}
	}

	return []string{}
}

func (curl *curl) getBodyArgument(tmpDir string) ([]string, error) {
	if len(curl.callData.Body) > 0 {
		tmpFile, err := curl.callData.GetBodyAsTempFile(tmpDir)

		if err != nil {
			return nil, err
		}

		curl.tmpFileName = tmpFile.Name()
		return []string{"-d", "@" + tmpFile.Name()}, nil
	}

	return []string{}, nil
}

func (curl *curl) runAsCmd(ctx context.Context) (Output, error) {
	args := []string{}
	for _, backendOpt := range curl.callData.BackendOptions {
		args = append(args, backendOpt...)
	}

	args = append(args, curl.getMethodArgument(false)...)
	for _, headerArgs := range curl.getHeaderArguments(false) {
		args = append(args, headerArgs...)
	}

	bodyArgs, err := curl.getBodyArgument("")
	if err != nil {
		return Output{}, err
	}

	args = append(args, bodyArgs...)
	args = append(args, curl.callData.Host.String())

	curlCmd := exec.CommandContext(ctx, curl.binaryName, args...)
	stdErrPipe, err := curlCmd.StderrPipe()
	if err != nil {
		return Output{}, err
	}
	stdOutPipe, err := curlCmd.StdoutPipe()
	if err != nil {
		return Output{}, err
	}
	curlCmd.Start()
	if err != nil {
		return Output{}, err
	}
	stdOut, err := io.ReadAll(stdOutPipe)
	if err != nil {
		return Output{StdOut: stdOut}, &BackedErr{
			Err:      err,
			ExitCode: curlCmd.ProcessState.ExitCode(),
		}
	}
	stdErr, err := io.ReadAll(stdErrPipe)
	if err != nil {
		return Output{StdOut: stdOut, StdErr: stdErr}, &BackedErr{
			Err:      err,
			ExitCode: curlCmd.ProcessState.ExitCode(),
		}
	}
	err = curlCmd.Wait()
	if err != nil {
		return Output{StdOut: stdOut, StdErr: stdErr}, &BackedErr{
			Err:      err,
			ExitCode: curlCmd.ProcessState.ExitCode(),
		}
	}
	return Output{StdOut: stdOut, StdErr: stdErr}, nil
}

func (curl *curl) getAsString() (string, error) {
	args := [][]string{}

	for _, optionLine := range curl.callData.BackendOptions {
		lineArguments := []string{}
		for _, option := range optionLine {
			lineArguments = append(lineArguments, utils.EscapeForShell(option))
		}
		args = append(args, lineArguments)
	}

	args = append(args, curl.getMethodArgument(true))
	args = append(args, curl.getHeaderArguments(true)...)

	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "Could not get current working dir, cannot store any body temp-file")
	}

	bodyArgs, err := curl.getBodyArgument(cwd)
	if err != nil {
		return "", err
	}

	args = append(args, bodyArgs)
	args = append(args, []string{
		utils.EscapeForShell(curl.callData.Host.String()),
	})

	output := "curl " + utils.PrettyPrintStringsForShell(args)

	return output, nil
}

func (curl *curl) cleanUp() error {
	if curl.tmpFileName != "" {
		return os.Remove(curl.tmpFileName)
	}

	return nil
}
