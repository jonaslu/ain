package parse

import (
	"context"
	"regexp"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/call"
)

type sourceMarker struct {
	lineContents    string
	sourceLineIndex int
}

func newSourceMarker(lineContents string, sourceLineIndex int) sourceMarker {
	return sourceMarker{lineContents: lineContents, sourceLineIndex: sourceLineIndex}
}

var emptyLine sourceMarker = newSourceMarker("emptyLine", -1)

func trimTemplate(template string) ([]sourceMarker, []string) {
	strippedLines := []sourceMarker{}

	templateLines := strings.Split(template, "\n")
	lastRowIndex := len(templateLines) - 1

	if lastRowIndex > 0 && templateLines[lastRowIndex] == "" {
		templateLines = templateLines[:len(templateLines)-1]
	}

	for sourceIndex, line := range templateLines {
		isCommentOrWhitespaceLine, _ := regexp.MatchString("^\\s*#|^\\s*$", line)
		if !isCommentOrWhitespaceLine {
			sourceMarker := sourceMarker{lineContents: strings.TrimSpace(line), sourceLineIndex: sourceIndex}
			strippedLines = append(strippedLines, sourceMarker)
		}
	}

	return strippedLines, templateLines
}

func ParseTemplate(ctx context.Context, template string) (*call.Data, []string) {
	var fatals []string

	trimmedTemplate, templateLines := trimTemplate(template)
	if len(trimmedTemplate) == 0 {
		return nil, []string{"Cannot process empty template"}
	}

	envVarsTemplate, envVarsFatals := transformEnvVars(trimmedTemplate)
	if len(envVarsFatals) > 0 {
		for _, transformFatalMarker := range envVarsFatals {
			fatals = append(fatals, formatFatalMarker(transformFatalMarker, templateLines))
		}

		return nil, fatals
	}

	shellCommandsTemplate, shellCommandFatals := transformShellCommands(ctx, envVarsTemplate)
	if len(shellCommandFatals) > 0 {
		for _, transformFatalMarker := range shellCommandFatals {
			fatals = append(fatals, formatFatalMarker(transformFatalMarker, templateLines))
		}

		return nil, fatals
	}

	callData := &call.Data{}
	if hostFatalMarker := parseHostSection(shellCommandsTemplate, callData); hostFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(hostFatalMarker, templateLines))
	}

	if headersFatalMarker := parseHeadersSection(shellCommandsTemplate, callData); headersFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(headersFatalMarker, templateLines))
	}

	if methodFatalMarker := parseMethodSection(shellCommandsTemplate, callData); methodFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(methodFatalMarker, templateLines))
	}

	if bodyFatalMarker := parseBodySection(shellCommandsTemplate, callData); bodyFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(bodyFatalMarker, templateLines))
	}

	return callData, fatals
}
