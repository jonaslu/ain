package utils

import "strings"

func UnsplitLineOnSeparator(commandArgs []string, unsplitSeparator string) []string {
	var unsplitLines []string
	var splitLine []string

	unsplitting := false

	for _, commandArg := range commandArgs {
		if strings.HasPrefix(commandArg, unsplitSeparator) {
			unsplitting = true
			commandArg = strings.TrimPrefix(commandArg, unsplitSeparator)
		}

		if strings.HasSuffix(commandArg, unsplitSeparator) {
			commandArg = strings.TrimSuffix(commandArg, unsplitSeparator)
			commandArg = strings.Join(splitLine, " ") + " " + commandArg

			unsplitting = false
			splitLine = nil
		}

		if unsplitting {
			splitLine = append(splitLine, commandArg)
		} else {
			unsplitLines = append(unsplitLines, commandArg)
		}
	}

	return unsplitLines
}
