package call

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

type curl struct {
	backendInput *data.BackendInput
	tmpFileName  string
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

func (curl *curl) getBodyArgument(tmpDir string) ([]string, error) {
	if len(curl.backendInput.Body) > 0 {
		tmpFile, err := curl.backendInput.GetBodyAsTempFile(tmpDir)

		if err != nil {
			return nil, err
		}

		curl.tmpFileName = tmpFile.Name()
		return []string{"-d", "@" + tmpFile.Name()}, nil
	}

	return []string{}, nil
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

	bodyArgs, err := curl.getBodyArgument("")
	if err != nil {
		return nil, err
	}

	args = append(args, bodyArgs...)
	args = append(args, curl.backendInput.Host.String())

	curlCmd := exec.CommandContext(ctx, curl.binaryName, args...)
	output, err := curlCmd.CombinedOutput()
	if err != nil {
		return output, &BackedErr{
			Err:      err,
			ExitCode: curlCmd.ProcessState.ExitCode(),
		}
	}

	return output, nil
}

func (curl *curl) getAsString() (string, error) {
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
		utils.EscapeForShell(curl.backendInput.Host.String()),
	})

	output := curl.binaryName + " " + utils.PrettyPrintStringsForShell(args)

	return output, nil
}

func (curl *curl) cleanUp() error {
	if curl.tmpFileName != "" {
		return os.Remove(curl.tmpFileName)
	}

	return nil
}
