package parse

import (
	"strconv"
	"strings"
)

type fatalMarker struct {
	message   string
	fatalLine sourceMarker
}

func newFatalMarker(message string, fatalLine sourceMarker) *fatalMarker {
	return &fatalMarker{message: message, fatalLine: fatalLine}
}

// Knows how to print errors and possibly an empty template into a string
func formatFatalMarker(fatalMarker *fatalMarker, templateLines []string) string {
	if fatalMarker.fatalLine == emptyLine {
		return "Fatal error " + fatalMarker.message
	}

	var templateContext []string

	errorLine := fatalMarker.fatalLine.sourceLineIndex
	lineBefore := errorLine - 1
	if lineBefore >= 0 {
		templateContext = append(templateContext, strconv.Itoa(lineBefore+1)+"   "+templateLines[lineBefore])
	}

	templateContext = append(templateContext, strconv.Itoa(errorLine+1)+" > "+templateLines[errorLine])

	lineAfter := errorLine + 1
	if lineAfter < len(templateLines) {
		templateContext = append(templateContext, strconv.Itoa(lineAfter+1)+"   "+templateLines[lineAfter])
	}

	message := "Fatal error " + fatalMarker.message + " on line " + strconv.Itoa(fatalMarker.fatalLine.sourceLineIndex+1) + ":\n"
	message = message + strings.Join(templateContext, "\n")

	return message
}
