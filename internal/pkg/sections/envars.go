package main

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

func (s *SectionedTemplate) substituteEnvVars() {
	for _, section := range s.sections {
		for idx := range *section {
			templateLine := &(*section)[idx]
			lineContents := templateLine.LineContents

			for _, envVarWithBrackets := range envVarExpressionRe.FindAllString(lineContents, -1) {
				envVarKeyStr := envVarKeyRe.FindStringSubmatch(envVarWithBrackets)
				if len(envVarKeyStr) != 2 {
					s.SetFatalMessage("Malformed variable", templateLine.SourceLineIndex)
					continue
				}

				envVarKey := envVarKeyStr[1]

				if envVarKey == "" {
					s.SetFatalMessage("Empty variable", templateLine.SourceLineIndex)
					continue
				}

				// I'll try anything that is not empty, if the user can't set (such as a variable with spaces in bash) it we can't find it anyway.
				// https://stackoverflow.com/questions/2821043/allowed-characters-in-linux-environment-variable-names
				value, exists := os.LookupEnv(envVarKey)

				if !exists {
					s.SetFatalMessage(formatMissingEnvVarErrorMessage(envVarKey), templateLine.SourceLineIndex)
				} else {
					if value == "" {
						s.SetFatalMessage(fmt.Sprintf("Value for variable %s is empty", envVarKey), templateLine.SourceLineIndex)
					} else {
						lineContents = strings.Replace(lineContents, envVarWithBrackets, value, 1)
					}
				}
			}

			templateLine.LineContents = lineContents
		}
	}
}