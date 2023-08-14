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

type httpie struct {
	backendInput *data.BackendInput
	tmpFileName  string
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

func (httpie *httpie) getBodyArgument(tmpDir string) (string, error) {
	tmpFile, err := httpie.backendInput.GetBodyAsTempFile(tmpDir)
	if err != nil {
		return "", err
	}

	httpie.tmpFileName = tmpFile.Name()
	return "@" + tmpFile.Name(), nil
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

	if len(httpie.backendInput.Body) > 0 {
		bodyArg, err := httpie.getBodyArgument("")

		if err != nil {
			return nil, err
		}

		args = append(args, bodyArg)
	}

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

func (httpie *httpie) getAsString() (string, error) {
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

	if len(httpie.backendInput.Body) > 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "Could not get current working dir, cannot store body temp-file")
		}

		bodyArg, err := httpie.getBodyArgument(cwd)
		if err != nil {
			return "", err
		}

		args = append(args, []string{bodyArg})
	}

	output := httpie.binaryName + " " + utils.PrettyPrintStringsForShell(args)

	return output, nil
}

func (httpie *httpie) cleanUp() error {
	if httpie.tmpFileName != "" {
		return os.Remove(httpie.tmpFileName)
	}

	return nil
}
