package parse

import (
	"fmt"
	"net/url"

	"github.com/jonaslu/ain/internal/pkg/call"
)

func parseHostSection(template []sourceMarker, callData *call.Data) *fatalMarker {
	captureResult, captureErr := captureSection("Host", template, true)
	if captureErr != nil {
		return captureErr
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return newFatalMarker("No mandatory [Host] section found", emptyLine)
	}

	var hostStr string
	for _, hostLine := range captureResult.sectionLines {
		hostStr = hostStr + hostLine.lineContents
	}

	host, err := url.Parse(hostStr)
	if err != nil {
		return newFatalMarker(
			fmt.Sprintf("[Host] has illegal url: %s, error: %v",
				hostStr,
				err),
			captureResult.sectionLines[0])
	}

	callData.Host = host

	return nil
}
