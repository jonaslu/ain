package call

import (
	"context"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

type httpie struct {
	backendInput *data.BackendInput
	binaryName   string
}

func prependIgnoreStdin(backendInput *data.BackendInput) {
	var foundIgnoreStdin bool

	for _, backendOptionLine := range backendInput.BackendOptions {
		for _, backendOption := range backendOptionLine {
			if backendOption == "--ignore-stdin" {
				foundIgnoreStdin = true
				break
			}
		}
	}

	if !foundIgnoreStdin {
		backendInput.BackendOptions = append([][]string{{"--ignore-stdin"}}, backendInput.BackendOptions...)
	}
}

func newHttpieBackend(backendInput *data.BackendInput, binaryName string) backend {
	prependIgnoreStdin(backendInput)
	return &httpie{
		backendInput: backendInput,
		binaryName:   binaryName,
	}
}

func (httpie *httpie) getMethodArgument() string {
	return strings.ToUpper(httpie.backendInput.Method)
}

func (httpie *httpie) getBodyArgument() []string {
	if httpie.backendInput.TempFileName != "" {
		return []string{"@" + httpie.backendInput.TempFileName}
	}

	return []string{}
}

func (httpie *httpie) runAsCmd(ctx context.Context) ([]byte, error) {
	args := []string{}
	for _, backendOpt := range httpie.backendInput.BackendOptions {
		args = append(args, backendOpt...)
	}

	if httpie.backendInput.Method != "" {
		args = append(args, httpie.getMethodArgument())
	}

	args = append(args, httpie.backendInput.Host.String())
	args = append(args, httpie.backendInput.Headers...)
	args = append(args, httpie.getBodyArgument()...)

	httpCmd := exec.CommandContext(ctx, httpie.binaryName, args...)
	output, err := httpCmd.CombinedOutput()

	if err != nil {
		return output, &BackedErr{
			Err:      err,
			ExitCode: httpCmd.ProcessState.ExitCode(),
		}
	}

	return output, nil
}

func (httpie *httpie) getAsString() string {
	args := [][]string{}
	for _, optionLine := range httpie.backendInput.BackendOptions {
		lineArguments := []string{}
		for _, option := range optionLine {
			lineArguments = append(lineArguments, utils.EscapeForShell(option))
		}
		args = append(args, lineArguments)
	}

	if httpie.backendInput.Method != "" {
		args = append(args, []string{utils.EscapeForShell(httpie.getMethodArgument())})
	}

	args = append(args, []string{utils.EscapeForShell(httpie.backendInput.Host.String())})

	for _, header := range httpie.backendInput.Headers {
		args = append(args, []string{utils.EscapeForShell(header)})
	}

	args = append(args, httpie.getBodyArgument())

	output := httpie.binaryName + " " + utils.PrettyPrintStringsForShell(args)

	return output
}
