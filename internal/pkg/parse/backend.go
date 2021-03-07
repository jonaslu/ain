package parse

import (
	"fmt"

	"github.com/jonaslu/ain/internal/pkg/call"
)

func parseBackendSection(template []sourceMarker, callData *call.Data) *fatalMarker {
	captureResult, captureFatal := captureSection("Backend", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return newFatalMarker("No mandatory [Backend] section found", emptyLine)
	}

	backendLines := captureResult.sectionLines

	if len(backendLines) > 1 {
		for _, backendLine := range backendLines {
			return newFatalMarker("Found several lines under [Backend]", backendLine)
		}
	}

	backendName := backendLines[0].lineContents

	if exists := callData.SetBackend(backendName); !exists {
		return newFatalMarker(fmt.Sprintf("Unknown backend: %s", backendName), backendLines[0])
	}

	return nil
}
