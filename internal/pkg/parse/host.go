package parse

import (
	"github.com/jonaslu/ain/internal/pkg/data"
)

func parseHostSection(template []sourceMarker, callData *data.Parse) *fatalMarker {
	captureResult, captureErr := captureSection("Host", template, true)
	if captureErr != nil {
		return captureErr
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	for _, hostLine := range captureResult.sectionLines {
		callData.Host = append(callData.Host, hostLine.lineContents)
	}

	return nil
}
