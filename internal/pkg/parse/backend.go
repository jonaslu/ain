package parse

import (
	"fmt"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

func (s *SectionedTemplate) getBackend() string {
	backendSourceMarkers := *s.GetNamedSection(BackendSection)
	if len(backendSourceMarkers) == 0 {
		return ""
	}

	if len(backendSourceMarkers) > 1 {
		s.SetFatalMessage("Found several lines under [Backend]", backendSourceMarkers[0].SourceLineIndex)
		return ""
	}

	backendSourceMarker := backendSourceMarkers[0]
	backend := strings.ToLower(backendSourceMarker.LineContents)

	if !call.ValidBackend(backend) {
		for backendName := range call.ValidBackends {
			if utils.LevenshteinDistance(backend, backendName) < 3 {
				s.SetFatalMessage(fmt.Sprintf("Unknown backend: %s. Did you mean %s", backend, backendName), backendSourceMarker.SourceLineIndex)
				return ""
			}
		}

		s.SetFatalMessage(fmt.Sprintf("Unknown backend %s", backend), backendSourceMarker.SourceLineIndex)
		return ""
	}

	return backend
}
