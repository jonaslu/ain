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

type executableAndArgs struct {
	executable string
	args       []string
}

type executableOutput struct {
	output       string
	fatalMessage string
}

func (s *sectionedTemplate) captureExecutableAndArgs() []executableAndArgs {
	executables := []executableAndArgs{}

	for _, sectionName := range sectionsAllowingExecutables {
		for _, templateLine := range *s.getNamedSection(sectionName) {
			lineContents := templateLine.LineContents

			for _, executableWithParens := range executableExpressionRe.FindAllString(lineContents, -1) {
				executableAndArgsCapture := executableRe.FindStringSubmatch(executableWithParens)

				if len(executableAndArgsCapture) != 2 {
					s.setFatalMessage("Malformed executable", templateLine.SourceLineIndex)
					continue
				}

				executableAndArgsStr := executableAndArgsCapture[1]
				if executableAndArgsStr == "" {
					s.setFatalMessage("Empty executable", templateLine.SourceLineIndex)
					continue
				}

				tokenizedExecutableLine, err := utils.TokenizeLine(executableAndArgsStr)
				if err != nil {
					s.setFatalMessage(err.Error(), templateLine.SourceLineIndex)
					continue
				}

				executable := tokenizedExecutableLine[0]

				executables = append(executables, executableAndArgs{
					executable: executable,
					args:       tokenizedExecutableLine[1:],
				})
			}
		}
	}

	return executables
}

func callExecutables(ctx context.Context, config data.Config, executables []executableAndArgs) []executableOutput {
	executableResults := make([]executableOutput, len(executables))

	wg := sync.WaitGroup{}
	for i, executable := range executables {
		go func(resultIndex int, executable executableAndArgs) {
			defer wg.Done()

			var stdout, stderr bytes.Buffer

			// !! TODO !! Bug right here, this is now enforced
			// per template and not for all templates.
			// It's also set a third time for the backends.
			// So in alles timeout*no templates + backend
			// I e waaay to looong
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

func (s *sectionedTemplate) insertExecutableOutput(executableResults *[]executableOutput) {
	for _, sectionHeader := range sectionsAllowingExecutables {
		section := s.getNamedSection(sectionHeader)
		anythingReplaced := false

		for idx := range *section {
			templateLine := &(*section)[idx]
			lineContents := templateLine.LineContents

			for _, executableWithParens := range executableExpressionRe.FindAllString(lineContents, -1) {
				result := (*executableResults)[0]
				*executableResults = (*executableResults)[1:]
				if result.fatalMessage != "" {
					s.setFatalMessage(result.fatalMessage, templateLine.SourceLineIndex)
					continue
				}

				lineContents = strings.Replace(lineContents, executableWithParens, result.output, 1)
				anythingReplaced = true
			}

			templateLine.LineContents = lineContents
		}

		if anythingReplaced {
			s.splitAndTrimSection(sectionHeader)
		}
	}
}
