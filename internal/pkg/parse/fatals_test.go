package parse

import "testing"

func Test_getLineWithNumberAndContent(t *testing.T) {
	tests := map[string]struct {
		lineIndex    int
		lineContents string
		addCaret     bool
		expected     string
	}{
		"Content and caret": {
			lineIndex:    1,
			lineContents: "test",
			addCaret:     true,
			expected:     "1 > test",
		},
		"Content and no caret": {
			lineIndex:    2,
			lineContents: "test",
			addCaret:     false,
			expected:     "2   test",
		},
		"No content no three space column": {
			lineIndex:    3,
			lineContents: "",
			addCaret:     true,
			expected:     "3",
		},
	}

	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			if got := getLineWithNumberAndContent(args.lineIndex, args.lineContents, args.addCaret); got != args.expected {
				t.Errorf("getLineWithNumberAndContent() = %v, want %v", got, args.expected)
			}
		})
	}
}
