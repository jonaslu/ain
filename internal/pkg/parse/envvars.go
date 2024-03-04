package parse

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/utils"
)

const maximumLevenshteinDistance = 2
const maximumNumberOfSuggestions = 3

var envVarExpressionRe = regexp.MustCompile(`(m?)\${[^}]*}?`)
var envVarKeyRe = regexp.MustCompile(`\${([^}]*)}`)

func formatMissingEnvVarErrorMessage(missingEnvVar string) string {
	suggestions := []string{}
	missingEnvVarLen := len(missingEnvVar)

	for _, envKeyValue := range os.Environ() {
		key := strings.SplitN(envKeyValue, "=", 2)[0]
		strLength := missingEnvVarLen - len(key)
		if strLength < 0 {
			strLength = -strLength
		}

		if strLength > maximumLevenshteinDistance {
			continue
		}

		if utils.LevenshteinDistance(missingEnvVar, key) <= maximumLevenshteinDistance {
			suggestions = append(suggestions, key)

			if len(suggestions) >= maximumNumberOfSuggestions {
				break
			}
		}
	}

	if len(suggestions) > 0 {
		return fmt.Sprintf("Cannot find value for variable %s. Did you mean %s", missingEnvVar, strings.Join(suggestions, " or "))
	}

	return fmt.Sprintf("Cannot find value for variable %s", missingEnvVar)
}

func (s *sectionedTemplate) substituteEnvVars() {
	newExpandedTemplateLines := []expandedSourceMarker{}

	for expandedTemplateLineIndex, expandedTemplateLine := range s.expandedTemplateLines {
		localExpandedTemplateLines := []expandedSourceMarker{}

		lineContents := expandedTemplateLine.LineContents
		noCommentsLineContents, _, _ := strings.Cut(expandedTemplateLine.LineContents, "#")

		foundEnvVarValues := []struct {
			value              string
			envVarWithBrackets string
		}{}

		prevNumFatals := len(s.fatals)

		for _, envVarWithBrackets := range envVarExpressionRe.FindAllString(noCommentsLineContents, -1) {
			envVarKeyStr := envVarKeyRe.FindStringSubmatch(envVarWithBrackets)
			if len(envVarKeyStr) != 2 {
				s.setFatalMessage("Malformed variable", expandedTemplateLineIndex)
				continue
			}

			envVarKey := envVarKeyStr[1]

			if envVarKey == "" {
				s.setFatalMessage("Empty variable", expandedTemplateLineIndex)
				continue
			}

			// I'll try anything that is not empty, if the user can't set (such as a variable with spaces in bash) it we can't find it anyway.
			// https://stackoverflow.com/questions/2821043/allowed-characters-in-linux-environment-variable-names
			value, exists := os.LookupEnv(envVarKey)

			if !exists {
				s.setFatalMessage(formatMissingEnvVarErrorMessage(envVarKey), expandedTemplateLineIndex)
				continue
			}

			if value == "" {
				s.setFatalMessage(fmt.Sprintf("Value for variable %s is empty", envVarKey), expandedTemplateLineIndex)
				continue
			}

			foundEnvVarValues = append(foundEnvVarValues, struct {
				value              string
				envVarWithBrackets string
			}{value, envVarWithBrackets})
		}

		if len(foundEnvVarValues) == 0 {
			newExpandedTemplateLines = append(newExpandedTemplateLines, expandedTemplateLine)
			continue
		}

		if prevNumFatals < len(s.fatals) {
			// When the fatal is set it reads from the expanded source
			// lines as it happens. The stuff below will never be read.
			continue
		}

		var expandedResult string
		tempLineContents := lineContents

		for _, foundValue := range foundEnvVarValues {
			foundIndex := strings.Index(tempLineContents, foundValue.envVarWithBrackets)

			expandedResult += tempLineContents[:foundIndex]
			expandedResult += foundValue.value

			tempLineContents = tempLineContents[foundIndex+len(foundValue.envVarWithBrackets):]
		}

		expandedResult += tempLineContents

		splitExpandedLines := strings.Split(strings.ReplaceAll(expandedResult, "\r\n", "\n"), "\n")

		for _, splitExpandedLine := range splitExpandedLines {
			localExpandedTemplateLines = append(localExpandedTemplateLines, expandedSourceMarker{
				sourceMarker: sourceMarker{
					LineContents:    splitExpandedLine,
					SourceLineIndex: expandedTemplateLine.SourceLineIndex,
				},
				expanded: true,
			})
		}

		newExpandedTemplateLines = append(newExpandedTemplateLines, localExpandedTemplateLines...)
	}

	s.expandedTemplateLines = newExpandedTemplateLines
}
