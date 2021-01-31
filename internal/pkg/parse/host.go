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

	hostLines := captureResult.sectionLines

	if len(hostLines) == 0 {
		return newFatalMarker("Empty [Host] line", captureResult.sectionHeaderLine)
	}

	if len(hostLines) > 1 {
		for _, hostLine := range hostLines {
			return newFatalMarker("Found several lines under [Host]", hostLine)
		}
	}

	hostLine := hostLines[len(hostLines)-1]
	hostStr := hostLine.lineContents
	host, err := url.Parse(hostStr)
	if err != nil {
		return newFatalMarker(fmt.Sprintf("Could not parse [Host] url: %v", captureErr), hostLine)
	}

	callData.Host = host

	return nil
}
