package parse

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

var executableExpressionRe = regexp.MustCompile(`(m?)\$\([^)]*\)?`)
var executableRe = regexp.MustCompile(`\$\(([^)]*)\)`)

var emptyOutputLineRe = regexp.MustCompile(`^\s*$`)

type executableAndArgs struct {
	executable string
	args       []string
}

type executableOutput struct {
	output       string
	fatalMessage string
}

func captureExecutableAndArgs(templateLines []sourceMarker) ([]executableAndArgs, []*fatalMarker) {
	var fatals []*fatalMarker
	executables := []executableAndArgs{}

	for _, templateLine := range templateLines {
		lineContents := templateLine.lineContents

		for _, executableWithParens := range executableExpressionRe.FindAllString(lineContents, -1) {
			executableAndArgsCapture := executableRe.FindStringSubmatch(executableWithParens)

			if len(executableAndArgsCapture) != 2 {
				fatals = append(fatals, newFatalMarker("Malformed executable", templateLine))
				continue
			}

			executableAndArgsStr := executableAndArgsCapture[1]
			if executableAndArgsStr == "" {
				fatals = append(fatals, newFatalMarker("Empty executable", templateLine))
				continue
			}

			tokenizedExecutableLine, err := utils.TokenizeLine(executableAndArgsStr)
			if err != nil {
				fatals = append(fatals, newFatalMarker(err.Error(), templateLine))
				continue
			}

			executable := tokenizedExecutableLine[0]

			executables = append(executables, executableAndArgs{
				executable: executable,
				args:       tokenizedExecutableLine[1:],
			})
		}
	}

	return executables, fatals
}

func callExecutables(ctx context.Context, config data.Config, executables []executableAndArgs) []executableOutput {
	executableResults := make([]executableOutput, len(executables))

	wg := sync.WaitGroup{}
	for i, executable := range executables {
		go func(resultIndex int, executable executableAndArgs) {
			defer wg.Done()

			var stdout, stderr bytes.Buffer

			timeoutCtx := ctx
			if config.Timeout != data.TimeoutNotSet {
				timeoutCtx, _ = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
			}

			cmd := exec.CommandContext(timeoutCtx, executable.executable, executable.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if timeoutCtx.Err() != nil {
				executableResults[resultIndex].fatalMessage = fmt.Sprintf("Executable %s timed out after %d seconds", cmd.String(), config.Timeout)
				return
			}

			stdoutStr := stdout.String()

			if err != nil {
				stderrStr := stderr.String()

				executableOutput := ""
				if stdoutStr != "" || stderrStr != "" {
					executableOutput = "\n" + strings.TrimSpace(strings.Join([]string{
						strings.TrimSpace(stdoutStr),
						strings.TrimSpace(stderrStr),
					}, " "))
				}

				executableResults[resultIndex].fatalMessage = fmt.Sprintf("Executable %s error: %v%s", cmd.String(), err, executableOutput)
				return
			}

			if stdoutStr == "" {
				executableResults[resultIndex].fatalMessage = fmt.Sprintf("Executable %s\nCommand produced no stdout output", cmd.String())
				return
			}

			executableResults[resultIndex].output = stdoutStr
		}(i, executable)

		wg.Add(1)
	}

	wg.Wait()

	return executableResults
}

func insertExecutableOutput(executableResults []executableOutput, templateLines []sourceMarker) ([]sourceMarker, []*fatalMarker) {
	var transformedTemplateLines []sourceMarker
	var fatals []*fatalMarker

	executableIndex := 0
	for _, templateLine := range templateLines {
		lineContents := templateLine.lineContents

		for _, executableWithParens := range executableExpressionRe.FindAllString(lineContents, -1) {
			result := executableResults[executableIndex]
			executableIndex++
			if result.fatalMessage != "" {
				fatals = append(fatals, newFatalMarker(result.fatalMessage, templateLine))
				continue
			}

			lineContents = strings.Replace(lineContents, executableWithParens, result.output, 1)
		}

		multilineOutput := strings.Split(strings.ReplaceAll(lineContents, "\r\n", "\n"), "\n")
		for _, lineOutput := range multilineOutput {
			if emptyOutputLineRe.MatchString(lineOutput) {
				continue
			}

			transformedTemplateLines = append(transformedTemplateLines, newSourceMarker(lineOutput, templateLine.sourceLineIndex))
		}
	}

	return transformedTemplateLines, fatals
}

func transformExecutables(ctx context.Context, config data.Config, templateLines []sourceMarker) ([]sourceMarker, []*fatalMarker) {
	executables, fatals := captureExecutableAndArgs(templateLines)
	if len(fatals) > 0 {
		return nil, fatals
	}

	executableResults := callExecutables(ctx, config, executables)

	return insertExecutableOutput(executableResults, templateLines)
}
