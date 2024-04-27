package parse

import (
	"fmt"
	"os"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/utils"
)

const maximumLevenshteinDistance = 2
const maximumNumberOfSuggestions = 3

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
	s.iterate(envVarToken, func(c token) (string, string) {
		envVarKey := c.content
		if envVarKey == "" {
			return "", "Empty variable"
		}

		// I'll try anything that is not empty, if the user can't set (such as a variable with spaces in bash) it we can't find it anyway.
		// https://stackoverflow.com/questions/2821043/allowed-characters-in-linux-environment-variable-names
		value, exists := os.LookupEnv(envVarKey)

		if !exists {
			return "", formatMissingEnvVarErrorMessage(envVarKey)
		}

		if value == "" {
			return "", fmt.Sprintf("Value for variable %s is empty", envVarKey)
		}

		return value, ""
	})
}
