package parse

import (
	"fmt"
	"strings"
)

type captureResult struct {
	sectionHeaderLine sourceMarker
	sectionLines      []sourceMarker
}

func captureSection(sectionName string, template []sourceMarker, trim bool) (*captureResult, *fatalMarker) {
	var sectionLines []sourceMarker
	sectionHeaderLine := emptyLine
	capturing := false

	for _, templateLine := range template {
		lineContents := templateLine.lineContents
		trimmedLineContents := strings.TrimSpace(templateLine.lineContents)

		if trimmedLineContents == "["+sectionName+"]" {
			if sectionHeaderLine != emptyLine {
				// !! TODO !! Capture all the places and make fatals accept several source-markers?
				return nil, newFatalMarker(fmt.Sprintf("Several [%s] sections found on line %d and %d", sectionName, sectionHeaderLine.sourceLineIndex, templateLine.sourceLineIndex), emptyLine)
			}

			sectionHeaderLine = templateLine
			capturing = true
			continue
		}

		if strings.HasPrefix(trimmedLineContents, "[") {
			capturing = false
			continue
		}

		if capturing {
			if trim {
				templateLine.lineContents = trimmedLineContents
			} else {
				templateLine.lineContents = lineContents
			}

			sectionLines = append(sectionLines, templateLine)
		}
	}

	if sectionHeaderLine != emptyLine && len(sectionLines) == 0 {
		return nil, newFatalMarker("Empty ["+sectionName+"] line", sectionHeaderLine)
	}

	captureResult := &captureResult{
		sectionHeaderLine: sectionHeaderLine,
		sectionLines:      sectionLines,
	}

	return captureResult, nil
}
