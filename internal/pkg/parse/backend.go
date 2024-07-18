package parse

import (
	"fmt"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

func (s *sectionedTemplate) getBackend() string {
	backendSourceMarkers := *s.getNamedSection(backendSection)
	if len(backendSourceMarkers) == 0 {
		return ""
	}

	if len(backendSourceMarkers) > 1 {
		s.setFatalMessage("Found several lines under [Backend]", backendSourceMarkers[0].sourceLineIndex)
		return ""
	}

	backendSourceMarker := backendSourceMarkers[0]
	backend := strings.ToLower(backendSourceMarker.lineContents)

	if !call.ValidBackend(backend) {
		for backendName := range call.ValidBackends {
			if utils.LevenshteinDistance(backend, backendName) < 3 {
				s.setFatalMessage(fmt.Sprintf("Unknown backend: %s. Did you mean %s", backend, backendName), backendSourceMarker.sourceLineIndex)
				return ""
			}
		}

		s.setFatalMessage(fmt.Sprintf("Unknown backend %s", backend), backendSourceMarker.sourceLineIndex)
		return ""
	}

	return backend
}
