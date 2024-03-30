package call

import (
	"context"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

type wget struct {
	backendInput *data.BackendInput
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

func (wget *wget) getBodyArgument() []string {
	if wget.backendInput.TempFileName != "" {
		return []string{"--body-file=" + wget.backendInput.TempFileName}
	}

	return []string{}
}

func (wget *wget) getAsCmd(ctx context.Context) *exec.Cmd {
	args := []string{}
	for _, backendOpt := range wget.backendInput.BackendOptions {
		args = append(args, backendOpt...)
	}

	if wget.backendInput.Method != "" {
		args = append(args, wget.getMethodArgument(false))
	}

	args = append(args, wget.getHeaderArguments(false)...)
	args = append(args, wget.getBodyArgument()...)

	args = append(args, wget.backendInput.Host.String())

	wgetCmd := exec.CommandContext(ctx, wget.binaryName, args...)
	return wgetCmd
}

func (wget *wget) getAsString() string {
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

	args = append(args, wget.getBodyArgument())

	args = append(args, []string{
		utils.EscapeForShell(wget.backendInput.Host.String()),
	})

	output := wget.binaryName + " " + utils.PrettyPrintStringsForShell(args)

	return output
}
