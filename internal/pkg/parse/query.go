package parse

import (
	"github.com/jonaslu/ain/internal/pkg/data"
)

func parseQuerySection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	sectionLines, captureFatal := captureSection("Query", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if len(sectionLines) == 0 {
		return nil
	}

	query := []string{}
	for _, queryLine := range sectionLines {
		query = append(query, queryLine.lineContents)
	}

	parsedTemplate.Query = query

	return nil
}
