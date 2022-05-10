package parse

import (
	"fmt"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

func parseBackendSection(template []sourceMarker, callData *data.Parse) *fatalMarker {
	captureResult, captureFatal := captureSection("Backend", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	backendLines := captureResult.sectionLines

	if len(backendLines) > 1 {
		for _, backendLine := range backendLines {
			return newFatalMarker("Found several lines under [Backend]", backendLine)
		}
	}

	requestedBackendName := strings.ToLower(backendLines[0].lineContents)

	if !call.ValidBackend(requestedBackendName) {
		for backendName, _ := range call.ValidBackends {
			if utils.LevenshteinDistance(requestedBackendName, backendName) < 3 {
				return newFatalMarker(fmt.Sprintf("Unknown backend: %s. Did you mean %s", requestedBackendName, backendName), backendLines[0])
			}
		}

		return newFatalMarker(fmt.Sprintf("Unknown backend %s", requestedBackendName), backendLines[0])
	}

	callData.Backend = requestedBackendName

	return nil
}
