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

	// !! TODO !! Empty template
	trimmedTemplate, templateLines := trimTemplate(template)

	// Transform

	// Parse in order of importance (URL, Headers, Body)
	// Concatenate any error and return them formatted with markers into the source

	callData := &call.Data{}
	hostFatalMarker := parseHostSection(trimmedTemplate, callData)
	if hostFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(hostFatalMarker, templateLines))
	}

	headersFatalMarker := parseHeadersSection(trimmedTemplate, callData)
	if headersFatalMarker != nil {
		fatals = append(fatals, formatFatalMarker(headersFatalMarker, templateLines))
	}

	return callData, fatals
}
