package parse

import "github.com/jonaslu/ain/internal/pkg/data"

func parseBodySection(template []sourceMarker, callData *data.Parse) *fatalMarker {
	captureResult, captureFatal := captureSection("Body", template, false)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	for _, bodyLineContents := range captureResult.sectionLines {
		callData.Body = append(callData.Body, bodyLineContents.lineContents)
	}

	return nil
}
