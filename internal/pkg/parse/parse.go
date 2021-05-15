package parse

import (
	"context"
	"regexp"
	"strings"
	"unicode"

	"github.com/jonaslu/ain/internal/pkg/data"
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
			sourceMarker := sourceMarker{
				lineContents:    strings.TrimRightFunc(line, func(r rune) bool { return unicode.IsSpace(r) }),
				sourceLineIndex: sourceIndex,
			}
			strippedLines = append(strippedLines, sourceMarker)
		}
	}

	return strippedLines, templateLines
}

func ParseTemplate(ctx context.Context, template string) (*data.Parse, []string) {
	var fatals []string

	trimmedTemplate, templateLines := trimTemplate(template)
	if len(trimmedTemplate) == 0 {
		return nil, []string{"Cannot process empty template"}
	}

	// !! TODO !! If this gets worse, put it in  it's on initializer method
	parseData := &data.Parse{}
	parseData.Config.Timeout = -1

	if configFatal := parseConfigSection(trimmedTemplate, parseData); configFatal != nil {
		return nil, []string{formatFatalMarker(configFatal, templateLines)}
	}

	envVarsTemplate, envVarsFatals := transformEnvVars(trimmedTemplate)
	if len(envVarsFatals) > 0 {
		for _, transformFatalMarker := range envVarsFatals {
			fatals = append(fatals, formatFatalMarker(transformFatalMarker, templateLines))
		}

		return nil, fatals
	}

	shellCommandsTemplate, shellCommandFatals := transformShellCommands(ctx, parseData.Config, envVarsTemplate)
	if len(shellCommandFatals) > 0 {
		for _, transformFatalMarker := range shellCommandFatals {
			fatals = append(fatals, formatFatalMarker(transformFatalMarker, templateLines))
		}

		return nil, fatals
	}

	sectionParsers := []func([]sourceMarker, *data.Parse) *fatalMarker{
		parseHostSection,
		parseHeadersSection,
		parseMethodSection,
		parseBodySection,
		parseBackendSection,
		parseBackendOptionsSection,
	}

	for _, sectionParser := range sectionParsers {
		if callFatalMarker := sectionParser(shellCommandsTemplate, parseData); callFatalMarker != nil {
			fatals = append(fatals, formatFatalMarker(callFatalMarker, templateLines))
		}
	}

	return parseData, fatals
}
