package parse

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

type executableAndArgs struct {
	executableCmd string
	args          []string
}

type executableOutput struct {
	cmdOutput    string
	fatalMessage string
}

func (s *sectionedTemplate) captureExecutableAndArgs() []executableAndArgs {
	executables := []executableAndArgs{}

	for expandedTemplateLineIndex, expandedTemplateLine := range s.expandedTemplateLines {
		for _, token := range expandedTemplateLine.tokens {
			if token.tokenType == commentToken {
				break
			}

			if token.tokenType != executableToken {
				continue
			}

			executableAndArgsStr := token.content
			if executableAndArgsStr == "" {
				s.setFatalMessage("Empty executable", expandedTemplateLineIndex)
				continue
			}

			tokenizedExecutableLine, err := utils.TokenizeLine(executableAndArgsStr)
			if err != nil {
				s.setFatalMessage(err.Error(), expandedTemplateLineIndex)
				continue
			}

			executable := tokenizedExecutableLine[0]

			executables = append(executables, executableAndArgs{
				executableCmd: executable,
				args:          tokenizedExecutableLine[1:],
			})
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

			cmd := exec.CommandContext(ctx, executable.executableCmd, executable.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if ctx.Err() == context.DeadlineExceeded {
				executableResults[resultIndex].fatalMessage = fmt.Sprintf("Executable %s timed out after %d seconds", cmd.String(), config.Timeout)
			}

			if ctx.Err() != nil {
				// Can't return an error, we're in a go-routine
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

			executableResults[resultIndex].cmdOutput = stdoutStr
		}(i, executable)

		wg.Add(1)
	}

	wg.Wait()

	return executableResults
}

func (s *sectionedTemplate) insertExecutableOutput(executableResults *[]executableOutput) {
	if len(*executableResults) == 0 {
		return
	}

	nextExecutableResult := (*executableResults)[0]

	s.iterate(executableToken, func(c token) (string, string) {
		fatalMessage := nextExecutableResult.fatalMessage
		output := nextExecutableResult.cmdOutput

		// > 1 because we have already processed the head of the list.
		// Hence at least two elements left, where the [1:] element is the
		// next item we're trying to consume.
		if len(*executableResults) > 1 {
			*executableResults = (*executableResults)[1:]
			nextExecutableResult = (*executableResults)[0]
		}

		return output, fatalMessage

	})
}
