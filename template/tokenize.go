package template

import (
	"regexp"
	"strings"
)

func TokenizeTemplate(editedTemplate string) Template {
	strippedLines := Template{}

	allLines := strings.Split(editedTemplate, "\n")
	for sourceIndex, line := range allLines {
		isCommentOrWhitespaceLine, _ := regexp.MatchString("^\\s*#|^\\s*$", line)
		if !isCommentOrWhitespaceLine {
			sourceMarker := SourceMarker{LineContents: strings.TrimSpace(line), SourceLineIndex: sourceIndex + 1}
			strippedLines = append(strippedLines, sourceMarker)
		}
	}

	return strippedLines
}
