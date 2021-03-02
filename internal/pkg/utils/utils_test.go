package utils

import (
	"reflect"
	"strings"
	"testing"
)

func tokenizeAndExpectError(commandLine string, t *testing.T) {
	_, err := TokenizeLine(commandLine, false)
	if err == nil {
		t.Fatal("Expected error to be non-nil")
	}

	if !strings.Contains(err.Error(), "Unterminated quote sequence") {
		t.Fatal("Expected unterminated quote sequence error message")
	}
}

func tokenizeAndCompareWithQuotes(commandLine string, expected []string, trimQuotes bool, t *testing.T) {
	res, err := TokenizeLine(commandLine, trimQuotes)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Result: %s, did not match expected: %s", strings.Join(res, "|"), strings.Join(expected, "|"))
	}
}

func TestEverythingPositive(t *testing.T) {
	commandLine := `""`
	tokenizeAndCompareWithQuotes(commandLine, []string{`""`}, false, t)

	commandLine = `" "`
	tokenizeAndCompareWithQuotes(commandLine, []string{`" "`}, false, t)

	commandLine = `"yaketi 😎 yak"`
	tokenizeAndCompareWithQuotes(commandLine, []string{`"yaketi 😎 yak"`}, false, t)

	commandLine = `''`
	tokenizeAndCompareWithQuotes(commandLine, []string{`''`}, false, t)

	commandLine = `   'doh'`
	tokenizeAndCompareWithQuotes(commandLine, []string{`'doh'`}, false, t)

	commandLine = `goat`
	tokenizeAndCompareWithQuotes(commandLine, []string{`goat`}, false, t)

	commandLine = `   goat`
	tokenizeAndCompareWithQuotes(commandLine, []string{`goat`}, false, t)

	commandLine = `"this has \" unquote"   `
	tokenizeAndCompareWithQuotes(commandLine, []string{`"this has \" unquote"`}, false, t)

	commandLine = `  'and this \' "" has unquote'`
	tokenizeAndCompareWithQuotes(commandLine, []string{`'and this \' "" has unquote'`}, false, t)

	commandLine = `\"`
	tokenizeAndCompareWithQuotes(commandLine, []string{`\"`}, false, t)

	commandLine = `"giat monkey  "`
	tokenizeAndCompareWithQuotes(commandLine, []string{`giat monkey`}, true, t)

	commandLine = `"giat '' monkey  "`
	tokenizeAndCompareWithQuotes(commandLine, []string{`giat '' monkey`}, true, t)

	commandLine = `""`
	tokenizeAndCompareWithQuotes(commandLine, nil, true, t)

	commandLine = `sheeba ""`
	tokenizeAndCompareWithQuotes(commandLine, []string{"sheeba"}, true, t)
}

func TestUnterminatedQuotes(t *testing.T) {
	tokenizeAndExpectError(`"`, t)
	tokenizeAndExpectError(`"\"`, t)
	tokenizeAndExpectError(`   "\"    `, t)

	tokenizeAndExpectError(`'`, t)
	tokenizeAndExpectError(`'\'`, t)
	tokenizeAndExpectError(`   '\'    `, t)
}
