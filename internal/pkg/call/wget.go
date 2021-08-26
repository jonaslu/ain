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

type wget struct {
	callData    *data.Call
	tmpFileName string
}

func prependOutputToStdin(callData *data.Call) {
	var foundOutputToStdin bool

	for _, backendOptionLine := range callData.BackendOptions {
		for _, backendOption := range backendOptionLine {
			// !! TODO !!
			// Use a regex here since there can be spaces between -O and -
			// -\a*O\s*-
			if backendOption == "-O-" {
				foundOutputToStdin = true
				break
			}
		}
	}

	if !foundOutputToStdin {
		callData.BackendOptions = append([][]string{{"-O-"}}, callData.BackendOptions...)
	}
}

func newWgetBackend(callData *data.Call) (*wget, error) {
	prependOutputToStdin(callData)
	returnValue := &wget{callData: callData}
	return returnValue, nil
}

func (wget *wget) getHeaderArguments(escape bool) []string {
	args := []string{}
	for _, header := range wget.callData.Headers {
		if escape {
			args = append(args, "--header="+utils.EscapeForShell(header))
		} else {
			args = append(args, "--header="+header)
		}
	}

	return args
}

func (wget *wget) getMethodArgument(escape bool) string {
	if wget.callData.Method != "" {
		methodCapitalized := strings.ToUpper(wget.callData.Method)

		if escape {
			return "--method=" + utils.EscapeForShell(methodCapitalized)
		}

		return "--method=" + methodCapitalized
	}

	return ""
}

func (wget *wget) getBodyArgument(tmpDir string) (string, error) {
	if len(wget.callData.Body) > 0 {
		tmpFile, err := wget.callData.GetBodyAsTempFile(tmpDir)

		if err != nil {
			return "", err
		}

		wget.tmpFileName = tmpFile.Name()
		return "--body-file=" + tmpFile.Name(), nil
	}

	return "", nil
}

func (wget *wget) runAsCmd(ctx context.Context) ([]byte, error) {
	args := []string{}
	for _, backendOpt := range wget.callData.BackendOptions {
		args = append(args, backendOpt...)
	}

	if wget.callData.Method != "" {
		args = append(args, wget.getMethodArgument(false))
	}

	args = append(args, wget.getHeaderArguments(false)...)

	if len(wget.callData.Body) > 0 {
		bodyArgs, err := wget.getBodyArgument("")
		if err != nil {
			return nil, err
		}

		args = append(args, bodyArgs)
	}

	args = append(args, wget.callData.Host.String())

	wgetCmd := exec.CommandContext(ctx, "wget", args...)
	output, err := wgetCmd.CombinedOutput()
	if err != nil {
		return output, &BackedErr{
			Err:      err,
			ExitCode: wgetCmd.ProcessState.ExitCode(),
		}
	}

	return output, nil
}

func (wget *wget) getAsString() (string, error) {
	args := [][]string{}

	for _, optionLine := range wget.callData.BackendOptions {
		lineArguments := []string{}
		for _, option := range optionLine {
			lineArguments = append(lineArguments, utils.EscapeForShell(option))
		}
		args = append(args, lineArguments)
	}

	args = append(args, []string{wget.getMethodArgument(true)})

	for _, header := range wget.getHeaderArguments(true) {
		args = append(args, []string{header})
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "Could not get current working dir, cannot store any body temp-file")
	}

	if len(wget.callData.Body) > 0 {
		bodyArg, err := wget.getBodyArgument(cwd)
		if err != nil {
			return "", err
		}

		args = append(args, []string{bodyArg})
	}
	args = append(args, []string{
		utils.EscapeForShell(wget.callData.Host.String()),
	})

	output := "wget " + utils.PrettyPrintStringsForShell(args)

	return output, nil
}

func (wget *wget) cleanUp() error {
	if wget.tmpFileName != "" {
		return os.Remove(wget.tmpFileName)
	}

	return nil
}
