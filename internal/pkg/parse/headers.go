package parse

import "github.com/jonaslu/ain/internal/pkg/call"

func parseHeadersSection(template []sourceMarker, callData *call.Data) *fatalMarker {
	captureResult, captureFatal := captureSection("Headers", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	headerLines := captureResult.sectionLines

	headers := []string{}
	findDuplicates := map[string]bool{}
	for _, headerLine := range headerLines {
		if _, exists := findDuplicates[headerLine.lineContents]; exists {
			return newFatalMarker("Same entry in [Headers] twice", headerLine)
		} else {
			findDuplicates[headerLine.lineContents] = true
			headers = append(headers, headerLine.lineContents)
		}
	}

	callData.Headers = headers

	return nil
}
