package call

import (
	"context"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

type curl struct {
	backendInput *data.BackendInput
	binaryName   string
}

func newCurlBackend(backendInput *data.BackendInput, binaryName string) backend {
	return &curl{
		backendInput: backendInput,
		binaryName:   binaryName,
	}
}

func (curl *curl) getHeaderArguments(escape bool) [][]string {
	args := [][]string{}
	for _, header := range curl.backendInput.Headers {
		headerVal := header
		if escape {
			headerVal = utils.EscapeForShell(header)
		}

		args = append(args, []string{"-H", headerVal})
	}

	return args
}

func (curl *curl) getMethodArgument(escape bool) []string {
	if curl.backendInput.Method != "" {
		methodCapitalized := strings.ToUpper(curl.backendInput.Method)
		if escape {
			methodCapitalized = utils.EscapeForShell(methodCapitalized)
		}

		return []string{"-X", methodCapitalized}
	}

	return []string{}
}

func (curl *curl) getBodyArgument() []string {
	if curl.backendInput.TempFileName != "" {
		return []string{"-d", "@" + curl.backendInput.TempFileName}
	}

	return []string{}
}

func (curl *curl) runAsCmd(ctx context.Context) ([]byte, error) {
	args := []string{}
	for _, backendOpt := range curl.backendInput.BackendOptions {
		args = append(args, backendOpt...)
	}

	args = append(args, curl.getMethodArgument(false)...)
	for _, headerArgs := range curl.getHeaderArguments(false) {
		args = append(args, headerArgs...)
	}

	args = append(args, curl.getBodyArgument()...)
	args = append(args, curl.backendInput.Host.String())

	curlCmd := exec.CommandContext(ctx, curl.binaryName, args...)
	output, err := curlCmd.CombinedOutput()
	if err != nil {
		return output, &BackedErr{
			Err:      err,
			ExitCode: curlCmd.ProcessState.ExitCode(),
		}
	}

	return output, err
}

func (curl *curl) getAsString() string {
	args := [][]string{}

	for _, optionLine := range curl.backendInput.BackendOptions {
		lineArguments := []string{}
		for _, option := range optionLine {
			lineArguments = append(lineArguments, utils.EscapeForShell(option))
		}
		args = append(args, lineArguments)
	}

	args = append(args, curl.getMethodArgument(true))
	args = append(args, curl.getHeaderArguments(true)...)

	args = append(args, curl.getBodyArgument())
	args = append(args, []string{
		utils.EscapeForShell(curl.backendInput.Host.String()),
	})

	output := curl.binaryName + " " + utils.PrettyPrintStringsForShell(args)

	return output
}
