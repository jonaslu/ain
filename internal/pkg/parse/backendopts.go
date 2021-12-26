package parse

import (
	"fmt"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/utils"
)

func parseBackendOptionsSection(template []sourceMarker, callData *data.Parse) *fatalMarker {
	captureResult, captureFatal := captureSection("BackendOptions", template, true)
	if captureFatal != nil {
		return captureFatal
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	for _, backendOptionLineContents := range captureResult.sectionLines {
		tokenizedBackendOpts, err := utils.TokenizeLine(backendOptionLineContents.lineContents)
		if err != nil {
			return newFatalMarker(fmt.Sprintf("Could not parse backend-option %s", err.Error()), backendOptionLineContents)
		}

		callData.BackendOptions = append(callData.BackendOptions, tokenizedBackendOpts)
	}

	return nil
}
