package sections

import (
	"fmt"
	"strings"

	"github.com/jonaslu/ain/template"
)

type captureResult struct {
	sectionHeaderLine template.SourceMarker
	sectionLines      template.Template
}

func captureSection(sectionName string, templ template.Template) (*captureResult, *Error) {
	var sectionLines template.Template
	sectionHeaderLine := template.EmptyLine
	capturing := false

	for _, templateLine := range templ {
		lineContents := templateLine.LineContents

		if lineContents == "["+sectionName+"]" {
			if sectionHeaderLine != template.EmptyLine {
				return nil, newError(fmt.Sprintf("Several [%s] sections found on line %d and %d", sectionName, sectionHeaderLine.SourceLineIndex, templateLine.SourceLineIndex), template.EmptyLine)
			}

			sectionHeaderLine = templateLine
			capturing = true
			continue
		}

		if strings.HasPrefix(lineContents, "[") {
			capturing = false
			continue
		}

		if capturing {
			sectionLines = append(sectionLines, templateLine)
		}
	}

	captureResult := &captureResult{
		sectionHeaderLine: sectionHeaderLine,
		sectionLines:      sectionLines,
	}

	return captureResult, nil
}
