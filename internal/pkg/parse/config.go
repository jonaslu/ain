package parse

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/pkg/errors"
)

var timeoutConfigRe = regexp.MustCompile(`(?i)\s*timeout\s*=\s*(-?\d+)?`)
var queryDelimRe = regexp.MustCompile(`(?i)\s*querydelim\s*=\s*(.*)`)

func parseQueryDelim(configStr string) (bool, string, error) {
	queryDelimMatch := queryDelimRe.FindStringSubmatch(configStr)
	if len(queryDelimMatch) != 2 {
		return false, "", nil
	}

	// We don't need to trim the string because it's already been
	// trimmed when converted into []sourceMarker:s
	queryDelim := queryDelimMatch[1]
	if strings.Contains(queryDelim, " ") {
		return true, "", errors.New("Delimiter cannot contain space")
	}

	return true, queryDelimMatch[1], nil
}

func parseTimeoutConfig(configStr string) (bool, int32, error) {
	timeoutMatch := timeoutConfigRe.FindStringSubmatch(configStr)
	if len(timeoutMatch) != 2 {
		return false, 0, nil
	}

	timeoutIntervalStr := timeoutMatch[1]
	if timeoutIntervalStr == "" {
		return true, 0, errors.New("Malformed timeout value, must be digit > 0")
	}

	timeoutIntervalInt64, err := strconv.ParseInt(timeoutIntervalStr, 10, 32)

	if err != nil {
		return true, 0, errors.Wrap(err, "Could not parse timeout [Config] interval")
	}

	if timeoutIntervalInt64 < 1 {
		return true, 0, errors.New("Timeout interval must be greater than 0")
	}

	return true, int32(timeoutIntervalInt64), nil
}

func parseConfigSection(template []sourceMarker, callData *data.Parse) *fatalMarker {
	captureResult, captureErr := captureSection("Config", template, true)
	if captureErr != nil {
		return captureErr
	}

	if captureResult.sectionHeaderLine == emptyLine {
		return nil
	}

	configLines := captureResult.sectionLines

	for _, configLine := range configLines {
		if isTimeoutConfig, timeoutValue, err := parseTimeoutConfig(configLine.lineContents); isTimeoutConfig {
			if callData.Config.Timeout > 0 {
				return newFatalMarker("Timeout config set twice", configLine)
			}

			if err != nil {
				return newFatalMarker(err.Error(), configLine)
			}

			callData.Config.Timeout = timeoutValue
			continue
		}

		if isQueryDelim, queryDelimValue, err := parseQueryDelim(configLine.lineContents); isQueryDelim {
			if callData.Config.QueryDelim != nil {
				return newFatalMarker("Query delimiter set twice", configLine)
			}

			if err != nil {
				return newFatalMarker(err.Error(), configLine)
			}

			callData.Config.QueryDelim = &queryDelimValue
			continue
		}
	}

	return nil
}
