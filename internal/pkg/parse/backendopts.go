package parse

import (
	"fmt"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

func parseBackendOptionsSection(template []sourceMarker, parsedTemplate *data.ParsedTemplate) *fatalMarker {
	sectionLines, captureFatal := captureSection("BackendOptions", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if len(sectionLines) == 0 {
		return nil
	}

	for _, backendOptionLineContents := range sectionLines {
		tokenizedBackendOpts, err := utils.TokenizeLine(backendOptionLineContents.lineContents)
		if err != nil {
			return newFatalMarker(fmt.Sprintf("Could not parse backend-option %s", err.Error()), backendOptionLineContents)
		}

		parsedTemplate.BackendOptions = append(parsedTemplate.BackendOptions, tokenizedBackendOpts)
	}

	return nil
}
