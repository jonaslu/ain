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
	callData    *data.Call
	tmpFileName string
	binaryName  string
}

func prependIgnoreStdin(callData *data.Call) {
	var foundIgnoreStdin bool

	for _, backendOptionLine := range callData.BackendOptions {
		for _, backendOption := range backendOptionLine {
			if backendOption == "--ignore-stdin" {
				foundIgnoreStdin = true
				break
			}
		}
	}

	if !foundIgnoreStdin {
		callData.BackendOptions = append([][]string{{"--ignore-stdin"}}, callData.BackendOptions...)
	}
}

func newHttpieBackend(callData *data.Call, binaryName string) backend {
	prependIgnoreStdin(callData)
	return &httpie{
		callData:   callData,
		binaryName: binaryName,
	}
}

func (httpie *httpie) getMethodArgument() string {
	return strings.ToUpper(httpie.callData.Method)
}

func (httpie *httpie) getBodyArgument(tmpDir string) (string, error) {
	tmpFile, err := httpie.callData.GetBodyAsTempFile(tmpDir)
	if err != nil {
		return "", err
	}

	httpie.tmpFileName = tmpFile.Name()
	return "@" + tmpFile.Name(), nil
}

func (httpie *httpie) generateCmd(ctx context.Context) (*exec.Cmd, error) {
	args := []string{}
	for _, backendOpt := range httpie.callData.BackendOptions {
		args = append(args, backendOpt...)
	}

	if httpie.callData.Method != "" {
		args = append(args, httpie.getMethodArgument())
	}

	args = append(args, httpie.callData.Host.String())
	args = append(args, httpie.callData.Headers...)

	if len(httpie.callData.Body) > 0 {
		bodyArg, err := httpie.getBodyArgument("")

		if err != nil {
			return nil, err
		}

		args = append(args, bodyArg)
	}
	httpCmd := exec.CommandContext(ctx, httpie.binaryName, args...)
        return httpCmd, nil
}

func (httpie *httpie) getAsString() (string, error) {
	args := [][]string{}
	for _, optionLine := range httpie.callData.BackendOptions {
		lineArguments := []string{}
		for _, option := range optionLine {
			lineArguments = append(lineArguments, utils.EscapeForShell(option))
		}
		args = append(args, lineArguments)
	}

	if httpie.callData.Method != "" {
		args = append(args, []string{utils.EscapeForShell(httpie.getMethodArgument())})
	}

	args = append(args, []string{utils.EscapeForShell(httpie.callData.Host.String())})

	for _, header := range httpie.callData.Headers {
		args = append(args, []string{utils.EscapeForShell(header)})
	}

	if len(httpie.callData.Body) > 0 {
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
