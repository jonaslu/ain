package sections

import (
	"github.com/jonaslu/ain/template"
)

func ParseHeadersSection(templ template.Template, templateSections *TemplateSections) (Warnings, *Error) {
	warnings := Warnings{}
	captureResult, captureErr := captureSection("Headers", templ, true)

	if captureErr != nil {
		return nil, captureErr
	}

	if captureResult.sectionHeaderLine == template.EmptyLine {
		return nil, nil
	}

	headerLines := captureResult.sectionLines

	headers := []string{}
	findDuplicates := map[string]bool{}
	for _, headerLine := range headerLines {
		if _, exists := findDuplicates[headerLine.LineContents]; exists {
			// !! TODO !! Maybe this should be an error instead?
			// Try it live and make it so if there is no apparent reason.
			warnings = addWarning(warnings, "Duplicate [Headers] entry", headerLine)
		} else {
			findDuplicates[headerLine.LineContents] = true
			headers = append(headers, headerLine.LineContents)
		}
	}

	templateSections.Headers = headers

	return warnings, nil
}
