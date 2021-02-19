package parse

import (
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

func ParseTemplate(template string) (*call.Data, []string) {
	var fatals []string

	trimmedTemplate, templateLines := trimTemplate(template)
	if len(trimmedTemplate) == 0 {
		return nil, []string{"Cannot process empty template"}
	}

	tranformedTemplate, transformFatals := transform(trimmedTemplate)
	if len(transformFatals) > 0 {
		for _, transformFatalMarker := range transformFatals {
			fatals = append(fatals, formatFatalMarker(transformFatalMarker, templateLines))
		}

		return nil, fatals
	}

	callData := &call.Data{}
	hostFatalMarker := parseHostSection(tranformedTemplate, callData)
	if hostFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(hostFatalMarker, templateLines))
	}

	headersFatalMarker := parseHeadersSection(tranformedTemplate, callData)
	if headersFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(headersFatalMarker, templateLines))
	}

	return callData, fatals
}
