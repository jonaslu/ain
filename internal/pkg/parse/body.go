package parse

import "github.com/jonaslu/ain/internal/pkg/data"

func parseBodySection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	captureResult, captureFatal := captureSection("Body", template, false)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	for _, bodyLineContents := range captureResult.sectionLines {
		parsedTemplate.Body = append(parsedTemplate.Body, bodyLineContents.lineContents)
	}

	return nil
}
