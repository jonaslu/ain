package sections

import (
	"fmt"
	"net/url"

	"github.com/jonaslu/ain/template"
)

func ParseHostSection(templ template.Template, templateSections *TemplateSections) (Warnings, *Error) {
	warnings := Warnings{}
	captureResult, captureErr := captureSection("Host", templ)

	if captureErr != nil {
		return nil, captureErr
	}

	if captureResult.sectionHeaderLine == template.EmptyLine {
		return nil, newError("No mandatory [Host] section found", template.EmptyLine)
	}

	hostLines := captureResult.sectionLines

	if len(hostLines) == 0 {
		return nil, newError("Empty [Host] line", captureResult.sectionHeaderLine)
	}

	if len(hostLines) > 1 {
		for _, hostLine := range hostLines {
			warnings = addWarning(warnings, "Found several lines under [Host]", hostLine)
		}
	}

	hostLine := hostLines[len(hostLines)-1]
	hostStr := hostLine.LineContents
	host, err := url.Parse(hostStr)
	if err != nil {
		return nil, newError(fmt.Sprintf("Could not parse [Host] url: %v", captureErr), hostLine)
	}

	templateSections.Host = host

	return warnings, nil
}
