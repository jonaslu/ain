package parse

import (
	"github.com/jonaslu/ain/internal/pkg/call"
)

func parseBackendOptionsSection(template []sourceMarker, callData *call.Data) *fatalMarker {
	captureResult, captureFatal := captureSection("BackendOptions", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	for _, backendOptionLineContents := range captureResult.sectionLines {
		callData.BackendOptions = append(callData.BackendOptions, backendOptionLineContents.lineContents)
	}

	return nil
}
