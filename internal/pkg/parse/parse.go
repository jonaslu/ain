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
var removeTrailingCommendRegExp = regexp.MustCompile("#.*$")
var isCommentOrWhitespaceRegExp = regexp.MustCompile(`^\s*#|^\s*$`)

func trimTemplate(template string) ([]sourceMarker, []string) {
	strippedLines := []sourceMarker{}

	templateLines := strings.Split(strings.ReplaceAll(template, "\r\n", "\n"), "\n")
	lastRowIndex := len(templateLines) - 1

	if lastRowIndex > 0 && templateLines[lastRowIndex] == "" {
		templateLines = templateLines[:len(templateLines)-1]
	}

	for sourceIndex, line := range templateLines {
		isCommentOrWhitespaceLine := isCommentOrWhitespaceRegExp.MatchString(line)
		if !isCommentOrWhitespaceLine {
			trailingCommentsRemoved := removeTrailingCommendRegExp.ReplaceAllString(line, "")

			sourceMarker := sourceMarker{
				lineContents:    strings.TrimRightFunc(trailingCommentsRemoved, func(r rune) bool { return unicode.IsSpace(r) }),
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

	envVarsTemplate, envVarsFatals := transformEnvVars(trimmedTemplate)
	if len(envVarsFatals) > 0 {
		for _, transformFatalMarker := range envVarsFatals {
			fatals = append(fatals, formatFatalMarker(transformFatalMarker, templateLines))
		}

		return nil, fatals
	}

	// !! TODO !! If this gets worse, put it in  it's on initializer method
	parseData := &data.Parse{}
	parseData.Config.Timeout = data.TimeoutNotSet

	if configFatal := parseConfigSection(envVarsTemplate, parseData); configFatal != nil {
		return nil, []string{formatFatalMarker(configFatal, templateLines)}
	}

	executablesTemplate, executableFatals := transformExecutables(ctx, parseData.Config, envVarsTemplate)
	if len(executableFatals) > 0 {
		for _, transformFatalMarker := range executableFatals {
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
		if callFatalMarker := sectionParser(executablesTemplate, parseData); callFatalMarker != nil {
			fatals = append(fatals, formatFatalMarker(callFatalMarker, templateLines))
		}
	}

	return parseData, fatals
}
