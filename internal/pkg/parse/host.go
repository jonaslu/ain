package parse

import (
	"github.com/jonaslu/ain/internal/pkg/data"
)

func parseHostSection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	captureResult, captureErr := captureSection("Host", template, true)
	if captureErr != nil {
		return captureErr
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	for _, hostLine := range captureResult.sectionLines {
		parsedTemplate.Host = append(parsedTemplate.Host, hostLine.lineContents)
	}

	return nil
}
