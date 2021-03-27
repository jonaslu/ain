package parse

import (
	"fmt"
	"net/url"

	"github.com/jonaslu/ain/internal/pkg/data"
)

func parseHostSection(template []sourceMarker, callData *data.Data) *fatalMarker {
	captureResult, captureErr := captureSection("Host", template, true)
	if captureErr != nil {
		return captureErr
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
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
