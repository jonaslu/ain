package call

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

type wget struct {
	backendInput *data.BackendInput
	tmpFileName  string
	binaryName   string
}

var outputToStdoutRegexp = regexp.MustCompile(`-\w*O\s*-`)

func prependOutputToStdin(backendInput *data.BackendInput) {
	var foundOutputToStdin bool

	for _, backendOptionLine := range backendInput.BackendOptions {
		backendOptions := strings.Join(backendOptionLine, " ")
		if outputToStdoutRegexp.MatchString(backendOptions) {
			foundOutputToStdin = true
			break
		}
	}

	if !foundOutputToStdin {
		backendInput.BackendOptions = append([][]string{{"-O-"}}, backendInput.BackendOptions...)
	}
}

func newWgetBackend(backendInput *data.BackendInput, binaryName string) backend {
	prependOutputToStdin(backendInput)
	return &wget{
		backendInput: backendInput,
		binaryName:   binaryName,
	}
}

func (wget *wget) getHeaderArguments(escape bool) []string {
	args := []string{}
	for _, header := range wget.backendInput.Headers {
		if escape {
			args = append(args, "--header="+utils.EscapeForShell(header))
		} else {
			args = append(args, "--header="+header)
		}
	}

	return args
}

func (wget *wget) getMethodArgument(escape bool) string {
	if wget.backendInput.Method != "" {
		methodCapitalized := strings.ToUpper(wget.backendInput.Method)

		if escape {
			return "--method=" + utils.EscapeForShell(methodCapitalized)
		}

		return "--method=" + methodCapitalized
	}

	return ""
}

func (wget *wget) getBodyArgument(tmpDir string) (string, error) {
	if len(wget.backendInput.Body) > 0 {
		tmpFile, err := wget.backendInput.GetBodyAsTempFile(tmpDir)

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
	for _, backendOpt := range wget.backendInput.BackendOptions {
		args = append(args, backendOpt...)
	}

	if wget.backendInput.Method != "" {
		args = append(args, wget.getMethodArgument(false))
	}

	args = append(args, wget.getHeaderArguments(false)...)

	if len(wget.backendInput.Body) > 0 {
		bodyArgs, err := wget.getBodyArgument("")
		if err != nil {
			return nil, err
		}

		args = append(args, bodyArgs)
	}

	args = append(args, wget.backendInput.Host.String())

	wgetCmd := exec.CommandContext(ctx, wget.binaryName, args...)
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

	for _, optionLine := range wget.backendInput.BackendOptions {
		lineArguments := []string{}
		for _, option := range optionLine {
			lineArguments = append(lineArguments, utils.EscapeForShell(option))
		}
		args = append(args, lineArguments)
	}

	if wget.backendInput.Method != "" {
		args = append(args, []string{wget.getMethodArgument(true)})
	}

	for _, header := range wget.getHeaderArguments(true) {
		args = append(args, []string{header})
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "Could not get current working dir, cannot store any body temp-file")
	}

	if len(wget.backendInput.Body) > 0 {
		bodyArg, err := wget.getBodyArgument(cwd)
		if err != nil {
			return "", err
		}

		args = append(args, []string{bodyArg})
	}

	args = append(args, []string{
		utils.EscapeForShell(wget.backendInput.Host.String()),
	})

	output := wget.binaryName + " " + utils.PrettyPrintStringsForShell(args)

	return output, nil
}

func (wget *wget) cleanUp() error {
	if wget.tmpFileName != "" {
		return os.Remove(wget.tmpFileName)
	}

	return nil
}
