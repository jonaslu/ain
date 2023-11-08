package parse

import (
	"fmt"

	"github.com/jonaslu/ain/internal/pkg/utils"
)

func (s *sectionedTemplate) getBackendOptions() [][]string {
	var backendOptions [][]string

	for _, backedOptionSourceMarker := range *s.getNamedSection(backendOptionsSection) {
		tokenizedBackendOpts, err := utils.TokenizeLine(backedOptionSourceMarker.LineContents)
		if err != nil {
			// !! TODO !! Can parse all messages don't have to return
			s.setFatalMessage(fmt.Sprintf("Could not parse backend-option %s", err.Error()), backedOptionSourceMarker.SourceLineIndex)
			return backendOptions
		}

		backendOptions = append(backendOptions, tokenizedBackendOpts)
	}

	return backendOptions
}
