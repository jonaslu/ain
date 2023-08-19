package parse

import (
	"github.com/jonaslu/ain/internal/pkg/data"
)

func parseHostSection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	sectionLines, captureErr := captureSection("Host", template, true)
	if captureErr != nil {
		return captureErr
	}

	if len(sectionLines) == 0 {
		return nil
	}

	for _, hostLine := range sectionLines {
		parsedTemplate.Host = append(parsedTemplate.Host, hostLine.lineContents)
	}

	return nil
}
