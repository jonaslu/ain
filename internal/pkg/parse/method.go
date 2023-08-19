package parse

import "github.com/jonaslu/ain/internal/pkg/data"

func parseMethodSection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	sectionLines, captureFatal := captureSection("Method", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if len(sectionLines) == 0 {
		return nil
	}

	if len(sectionLines) > 1 {
		for _, hostLine := range sectionLines {
			return newFatalMarker("Found several lines under [Method]", hostLine)
		}
	}

	parsedTemplate.Method = sectionLines[0].lineContents

	return nil
}
