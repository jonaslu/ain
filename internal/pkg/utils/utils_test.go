package utils

import (
	"reflect"
	"testing"
)

func TestTokenizeLineGoodCases(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected []string
	}{
		`discard spaces`: {
			`   word   `,
			[]string{`word`},
		},

		`collect last word in buffer`: {
			`word`,
			[]string{`word`},
		},

		`split on two words`: {
			`word1 word2`,
			[]string{`word1`, `word2`},
		},

		`escaped quote`: {
			`wo\"rd`,
			[]string{`wo"rd`},
		},

		`escaped quote when quoting`: {
			`"wo\"rd"`,
			[]string{`wo"rd`},
		},

		`other quote when quoting`: {
			`"wo\'rd"`,
			[]string{`wo\'rd`},
		},

		`quote at edge`: {
			`word\"`,
			[]string{`word"`},
		},

		`quoting in the middle`: {
			`wo"  rd  "`,
			[]string{`wo  rd  `},
		},

		`quote and quote back to back`: {
			`"word""word"`,
			[]string{`wordword`},
		},

		`quoted whitespace in the beginning and end should be retained`: {
			`"   word    "`,
			[]string{`   word    `},
		},
	}

	for name, test := range tests {
		res, err := TokenizeLine(test.input)
		if err != nil {
			t.Fatalf("Got unexpected error: %v", err)
		}

		if !reflect.DeepEqual(test.expected, res) {
			t.Fatalf("Test: %s, got: %v, expected: %v", name, res, test.expected)
		}
	}
}

func TestTokenizeLineBadCases(t *testing.T) {
	tests := map[string]struct {
		input        string
		errorMessage string
	}{
		`single quote`: {
			`"`,
			`Unterminated quote sequence: "`,
		},

		`beginning of word`: {
			`"word`,
			`Unterminated quote sequence: "word`,
		},

		`end of word`: {
			`word"`,
			`Unterminated quote sequence: word"`,
		},

		`maximum context with no ellipsize`: {
			`the'eht`,
			`Unterminated quote sequence: the'eht`,
		},

		`ellipsize with three chars of context on each side of the quote`: {
			`this is a quote'with no termination`,
			`Unterminated quote sequence: ...ote'wit...`,
		},
	}

	for name, test := range tests {
		_, err := TokenizeLine(test.input)
		if err == nil {
			t.Fatal("No expected error")
		}

		if err.Error() != test.errorMessage {
			t.Fatalf("Test: %s, got error message: %v, expected: %v", name, err.Error(), test.errorMessage)
		}
	}
}

func TestEllipsize(t *testing.T) {
	tests := map[string]struct {
		from     int
		to       int
		input    string
		expected string
	}{
		"indexes out of bounds, empty string": {
			from:     -1,
			to:       1,
			input:    "",
			expected: "",
		},
		"indexes out of bounds, valid string": {
			from:     -1,
			to:       2,
			input:    "a",
			expected: "a",
		},
		"maximum length not ellipsized": {
			from:     3,
			to:       4,
			input:    "abcdefg",
			expected: "abcdefg",
		},
		"left side ellipsized": {
			from:     4,
			to:       5,
			input:    "abcdefgh",
			expected: "...efgh",
		},
		"right side ellipsized": {
			from:     3,
			to:       4,
			input:    "abcdefgh",
			expected: "abcd...",
		},
	}

	for name, test := range tests {
		if res := Ellipsize(test.from, test.to, test.input); res != test.expected {
			t.Fatalf("Test: %s, got: %v, expected: %v", name, res, test.expected)
		}
	}
}
