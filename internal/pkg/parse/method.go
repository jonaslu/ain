package parse

import (
	"github.com/jonaslu/ain/internal/pkg/call"
)

func parseMethodSection(template []sourceMarker, callData *call.Data) *fatalMarker {
	captureResult, captureFatal := captureSection("Method", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	methodLines := captureResult.sectionLines

	if len(methodLines) > 1 {
		for _, hostLine := range methodLines {
			return newFatalMarker("Found several lines under [Method]", hostLine)
		}
	}

	callData.Method = methodLines[0].lineContents

	return nil
}
