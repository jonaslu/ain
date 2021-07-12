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

var subShellExpressionRe = regexp.MustCompile(`(m?)\$\([^)]*\)?`)
var subShellCommandRe = regexp.MustCompile(`\$\(([^)]*)\)`)

type shellCommandAndArgs struct {
	cmd  string
	args []string
}

type shellCommandOutput struct {
	output       string
	fatalMessage string
}

func captureShellCommandAndArgs(templateLines []sourceMarker) ([]shellCommandAndArgs, []*fatalMarker) {
	var fatals []*fatalMarker
	shellCommands := []shellCommandAndArgs{}

	for _, templateLine := range templateLines {
		lineContents := templateLine.lineContents

		for _, subShellCallWithParens := range subShellExpressionRe.FindAllString(lineContents, -1) {
			shellCommandAndArgsCapture := subShellCommandRe.FindStringSubmatch(subShellCallWithParens)

			if len(shellCommandAndArgsCapture) != 2 {
				fatals = append(fatals, newFatalMarker("Malformed shell command", templateLine))
				continue
			}

			shellCommandAndArgsStr := shellCommandAndArgsCapture[1]
			if shellCommandAndArgsStr == "" {
				fatals = append(fatals, newFatalMarker("Empty shell command", templateLine))
				continue
			}

			tokenizedCommandLine, err := utils.TokenizeLine(shellCommandAndArgsStr, true)
			if err != nil {
				fatals = append(fatals, newFatalMarker(err.Error(), templateLine))
				continue
			}

			command := tokenizedCommandLine[0]

			shellCommands = append(shellCommands, shellCommandAndArgs{
				cmd:  command,
				args: tokenizedCommandLine[1:],
			})
		}
	}

	return shellCommands, fatals
}

func callShellCommands(ctx context.Context, config data.Config, shellCommands []shellCommandAndArgs) []shellCommandOutput {
	shellResults := make([]shellCommandOutput, len(shellCommands))

	wg := sync.WaitGroup{}
	for i, shellCommand := range shellCommands {
		go func(resultIndex int, shellCommand shellCommandAndArgs) {
			defer wg.Done()

			var stdout, stderr bytes.Buffer

			timeoutCtx := ctx
			if config.Timeout > -1 {
				timeoutCtx, _ = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
			}

			cmd := exec.CommandContext(timeoutCtx, shellCommand.cmd, shellCommand.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if timeoutCtx.Err() != nil {
				shellResults[resultIndex].fatalMessage = fmt.Sprintf("Command: %s timed out after %d seconds ", cmd.String(), config.Timeout)
				return
			}

			stdoutStr := stdout.String()

			if err != nil {
				stderrStr := stderr.String()
				shellResults[resultIndex].fatalMessage = fmt.Sprintf("Error: %v running command: %s. Command output: %s %s", err, cmd.String(), stderrStr, stdoutStr)
				return
			}

			if stdoutStr == "" {
				shellResults[resultIndex].fatalMessage = fmt.Sprintf("Error running command: %s. Command produced no stdout output", cmd.String())
				return
			}

			shellResults[resultIndex].output = stdoutStr
		}(i, shellCommand)

		wg.Add(1)
	}

	wg.Wait()

	return shellResults
}

func insertShellCommandOutput(shellResults []shellCommandOutput, templateLines []sourceMarker) ([]sourceMarker, []*fatalMarker) {
	var transformedTemplateLines []sourceMarker
	var fatals []*fatalMarker

	subShellIndex := 0
	for _, templateLine := range templateLines {
		lineContents := templateLine.lineContents

		for _, subShellCallWithParens := range subShellExpressionRe.FindAllString(lineContents, -1) {
			result := shellResults[subShellIndex]
			subShellIndex++
			if result.fatalMessage != "" {
				fatals = append(fatals, newFatalMarker(result.fatalMessage, templateLine))
				continue
			}

			lineContents = strings.Replace(lineContents, subShellCallWithParens, result.output, 1)
		}

		multilineOutput := strings.Split(strings.ReplaceAll(lineContents, "\r\n", "\n"), "\n")
		for _, lineOutput := range multilineOutput {
			transformedTemplateLines = append(transformedTemplateLines, newSourceMarker(lineOutput, templateLine.sourceLineIndex))
		}
	}

	return transformedTemplateLines, fatals
}

func transformShellCommands(ctx context.Context, config data.Config, templateLines []sourceMarker) ([]sourceMarker, []*fatalMarker) {
	shellCommands, fatals := captureShellCommandAndArgs(templateLines)
	if len(fatals) > 0 {
		return nil, fatals
	}

	shellResults := callShellCommands(ctx, config, shellCommands)

	return insertShellCommandOutput(shellResults, templateLines)
}
