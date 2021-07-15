package parse

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var envVarExpressionRe = regexp.MustCompile(`(m?)\${[^}]*}?`)
var envVarValueRe = regexp.MustCompile(`\${([^}]*)}`)

func transformEnvVars(templateLines []sourceMarker) ([]sourceMarker, []*fatalMarker) {
	var fatals []*fatalMarker
	var transformedTemplateLines []sourceMarker

	for _, templateLine := range templateLines {
		lineContents := templateLine.lineContents

		for _, envVarWithBrackets := range envVarExpressionRe.FindAllString(lineContents, -1) {
			envVarValues := envVarValueRe.FindStringSubmatch(envVarWithBrackets)
			if len(envVarValues) != 2 {
				fatals = append(fatals, newFatalMarker("Malformed variable", templateLine))
				continue
			}

			envVarValue := envVarValues[1]

			if envVarValue == "" {
				fatals = append(fatals, newFatalMarker("Empty variable", templateLine))
				continue
			}

			// I'll try anything that is not empty, if the user can't set (such as a variable with spaces in bash) it we can't find it anyway.
			// https://stackoverflow.com/questions/2821043/allowed-characters-in-linux-environment-variable-names
			value, exists := os.LookupEnv(envVarValue)

			if !exists {
				fatals = append(fatals, newFatalMarker(fmt.Sprintf("Cannot find value for variable %s", envVarValue), templateLine))
			} else {
				if value == "" {
					fatals = append(fatals, newFatalMarker(fmt.Sprintf("Value for variable %s is empty", envVarValue), templateLine))
				} else {
					lineContents = strings.Replace(lineContents, envVarWithBrackets, value, 1)
				}
			}
		}

		transformedTemplateLines = append(transformedTemplateLines, newSourceMarker(lineContents, templateLine.sourceLineIndex))
	}

	return transformedTemplateLines, fatals
}
