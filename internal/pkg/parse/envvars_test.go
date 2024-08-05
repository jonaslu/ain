package parse

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_sectionedTemplate_expandEnvVars2_GoodCases(t *testing.T) {
	tests := map[string]struct {
		beforeTest     func()
		inputTemplate  string
		expectedResult []expandedSourceMarker
	}{
		"Substitution works": {
			beforeTest: func() {
				os.Setenv("VAR1", "value1")
				os.Setenv("VAR2", "value2")
			},
			inputTemplate: "${VAR1} ${VAR2}",
			expectedResult: []expandedSourceMarker{{
				tokens:          nil,
				content:         "value1 value2",
				comment:         "",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
	}
	for name, test := range tests {
		test.beforeTest()
		s := newSectionedTemplate2(test.inputTemplate, "testing")

		if s.substituteEnvVars2(); s.hasFatalMessages() {
			t.Errorf("Got unexpected fatals, %s ", s.getFatalMessages())
		} else {
			if !reflect.DeepEqual(test.expectedResult, s.expandedTemplateLines) {
				t.Errorf("Test: %s. Expected %v, got: %v", name, test.expectedResult, s.expandedTemplateLines)
			}
		}
	}
}

func Test_sectionedTemplate_expandEnvVars2_BadCases(t *testing.T) {
	tests := map[string]struct {
		beforeTest           func()
		input                string
		expectedErrorMessage string
	}{
		"Empty variable": {
			beforeTest:           func() {},
			input:                "${}",
			expectedErrorMessage: "Empty variable",
		},
		"Cannot find value for variable": {
			beforeTest: func() {
				os.Unsetenv("VAR")
			},
			input:                "${VAR}",
			expectedErrorMessage: "Cannot find value for variable VAR",
		},
		"Value for variable is empty": {
			beforeTest: func() {
				os.Setenv("VAR", "")
			},
			input:                "${VAR}",
			expectedErrorMessage: "Value for variable VAR is empty",
		},
	}

	for name, test := range tests {
		test.beforeTest()
		s := newSectionedTemplate2(test.input, "")
		s.substituteEnvVars2()

		if len(s.fatals) != 1 {
			t.Errorf("Test: %s. Wrong number of fatals", name)
		}

		if !strings.Contains(s.fatals[0], test.expectedErrorMessage) {
			t.Errorf("Test: %s. Unexpected error message: %s", name, s.fatals[0])
		}
	}
}
