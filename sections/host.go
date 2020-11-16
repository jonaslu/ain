package sections

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jonaslu/ain/template"
)

func ParseHostSection(templ template.Template, templateSections *TemplateSections) *ParseResult {
	parseResult := &ParseResult{}

	var hostLines template.Template
	hostSectionLine := template.EmptyLine
	capturing := false

	for _, templateLine := range templ {
		lineContents := templateLine.LineContents

		if lineContents == "[Host]" {
			if hostSectionLine != template.EmptyLine {
				parseResult.addError(fmt.Sprintf("Several [Host] sections found on line %d and %d", hostSectionLine.SourceLineIndex, templateLine.SourceLineIndex), template.EmptyLine)
				return parseResult
			}

			hostSectionLine = templateLine
			capturing = true
			continue
		}

		if strings.HasPrefix(lineContents, "[") {
			capturing = false
			continue
		}

		if capturing {
			hostLines = append(hostLines, templateLine)
		}
	}

	if hostSectionLine == template.EmptyLine {
		parseResult.addError("No mandatory [Host] section found", template.EmptyLine)
		return parseResult
	}

	if len(hostLines) == 0 {
		parseResult.addError("Empty [Host] line", hostSectionLine)
		return parseResult
	}

	if len(hostLines) > 1 {
		for _, hostLine := range hostLines {
			parseResult.addWarning("Found several host lines", hostLine)
		}
	}

	hostLine := hostLines[len(hostLines)-1]
	hostStr := hostLine.LineContents
	host, err := url.Parse(hostStr)
	if err != nil {
		parseResult.addError(fmt.Sprintf("Could not parse [Host] url: %v", err), hostLine)
	}

	templateSections.Host = host

	return parseResult
}
