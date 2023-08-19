package parse

import "github.com/jonaslu/ain/internal/pkg/data"

func parseBodySection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	sectionLines, captureFatal := captureSection("Body", template, false)
	if captureFatal != nil {
		return captureFatal
	}

	if len(sectionLines) == 0 {
		return nil
	}

	for _, bodyLineContents := range sectionLines {
		parsedTemplate.Body = append(parsedTemplate.Body, bodyLineContents.lineContents)
	}

	return nil
}
