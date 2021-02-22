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
)

// !! TODO !! This should be a global config things
const cmdTimeOutInSeconds = 10

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

func unsplitLineOnSeparator(commandArgs []string, unsplitSeparator string) []string {
	var unsplitLines []string
	var splitLine []string

	unsplitting := false

	for _, commandArg := range commandArgs {
		if strings.HasPrefix(commandArg, unsplitSeparator) {
			unsplitting = true
			commandArg = strings.TrimPrefix(commandArg, unsplitSeparator)
		}

		if strings.HasSuffix(commandArg, unsplitSeparator) {
			commandArg = strings.TrimSuffix(commandArg, unsplitSeparator)
			commandArg = strings.Join(splitLine, " ") + " " + commandArg

			unsplitting = false
			splitLine = nil
		}

		if unsplitting {
			splitLine = append(splitLine, commandArg)
		} else {
			unsplitLines = append(unsplitLines, commandArg)
		}
	}

	return unsplitLines
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
			shellCommandAndArgsSlice := strings.Split(shellCommandAndArgsStr, " ")

			command := shellCommandAndArgsSlice[0]

			if command == "" {
				fatals = append(fatals, newFatalMarker("Empty shell command", templateLine))
				continue
			}

			args := shellCommandAndArgsSlice[1:]
			shellCommandAndArgsSlice = unsplitLineOnSeparator(args, "\"")
			shellCommandAndArgsSlice = unsplitLineOnSeparator(shellCommandAndArgsSlice, "'")

			shellCommands = append(shellCommands, shellCommandAndArgs{
				cmd:  command,
				args: shellCommandAndArgsSlice,
			})
		}
	}

	return shellCommands, fatals
}

func callShellCommands(ctx context.Context, shellCommands []shellCommandAndArgs) []shellCommandOutput {
	shellResults := make([]shellCommandOutput, len(shellCommands))

	wg := sync.WaitGroup{}
	for i, shellCommand := range shellCommands {
		go func(resultIndex int, shellCommand shellCommandAndArgs) {
			defer wg.Done()

			var stdout, stderr bytes.Buffer
			timeoutCtx, _ := context.WithTimeout(ctx, cmdTimeOutInSeconds*time.Second)
			cmd := exec.CommandContext(timeoutCtx, shellCommand.cmd, shellCommand.args...)

			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if timeoutCtx.Err() != nil {
				shellResults[resultIndex].fatalMessage = fmt.Sprintf("Command: %s timed out after %d seconds ", cmd.String(), cmdTimeOutInSeconds)
				return
			}

			stdoutStr := string(stdout.Bytes())
			stderrStr := string(stdout.Bytes())

			if err != nil {
				shellResults[resultIndex].fatalMessage = fmt.Sprintf("Error: %v running command: %s. Command stderr: %s %s", err, cmd.String(), stderrStr, stdoutStr)
				return
			}

			if stdoutStr == "" {
				shellResults[resultIndex].fatalMessage = fmt.Sprintf("Error running command: %s. Command produced no stdout output", cmd.String())
				return
			}

			shellResults[resultIndex].output = string(stdout.Bytes())
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

		transformedTemplateLines = append(transformedTemplateLines, newSourceMarker(lineContents, templateLine.sourceLineIndex))
	}

	return transformedTemplateLines, fatals
}

func transformShellCommands(ctx context.Context, templateLines []sourceMarker) ([]sourceMarker, []*fatalMarker) {
	shellCommands, fatals := captureShellCommandAndArgs(templateLines)
	if len(fatals) > 0 {
		return nil, fatals
	}

	shellResults := callShellCommands(ctx, shellCommands)

	return insertShellCommandOutput(shellResults, templateLines)
}
