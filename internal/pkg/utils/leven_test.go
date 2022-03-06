package utils

import (
	"testing"
)

func Test_LevenshteinDistance(t *testing.T) {
	tests := map[string]struct {
		str1         string
		str2         string
		expectedCost int
	}{
		"empty string to empty string": {
			"",
			"",
			0,
		},
		"empty string to str2": {
			"",
			"abc",
			3,
		},
		"empty string to str1": {
			"abc",
			"",
			3,
		},
		"one insertion": {
			"a",
			"ab",
			1,
		},
		"one deletion": {
			"ab",
			"a",
			1,
		},
		"one change": {
			"a",
			"b",
			1,
		},
	}

	for testCase, testData := range tests {
		actualCost := LevenshteinDistance(testData.str1, testData.str2)
		if actualCost != testData.expectedCost {
			t.Errorf("Cost of %s failed, got: %d expected: %d", testCase, actualCost, testData.expectedCost)
			t.Fail()
		}
	}
}
