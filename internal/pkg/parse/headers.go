package parse

import "github.com/jonaslu/ain/internal/pkg/data"

func parseHeadersSection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	sectionLines, captureFatal := captureSection("Headers", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if len(sectionLines) == 0 {
		return nil
	}

	headers := []string{}
	for _, headerLine := range sectionLines {
		headers = append(headers, headerLine.lineContents)
	}

	parsedTemplate.Headers = headers

	return nil
}
