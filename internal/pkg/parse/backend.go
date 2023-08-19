package parse

import (
	"fmt"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

func parseBackendSection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	sectionLines, captureFatal := captureSection("Backend", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if len(sectionLines) == 0 {
		return nil
	}

	if len(sectionLines) > 1 {
		for _, backendLine := range sectionLines {
			return newFatalMarker("Found several lines under [Backend]", backendLine)
		}
	}

	requestedBackendName := strings.ToLower(sectionLines[0].lineContents)

	if !call.ValidBackend(requestedBackendName) {
		for backendName, _ := range call.ValidBackends {
			if utils.LevenshteinDistance(requestedBackendName, backendName) < 3 {
				return newFatalMarker(fmt.Sprintf("Unknown backend: %s. Did you mean %s", requestedBackendName, backendName), sectionLines[0])
			}
		}

		return newFatalMarker(fmt.Sprintf("Unknown backend %s", requestedBackendName), sectionLines[0])
	}

	parsedTemplate.Backend = requestedBackendName

	return nil
}
