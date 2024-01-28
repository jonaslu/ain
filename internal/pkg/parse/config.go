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

func (s *sectionedTemplate) getConfig() data.Config {
	config := data.NewConfig()

	for _, configLine := range *s.getNamedSection(configSection) {
		if isTimeoutConfig, timeoutValue, err := parseTimeoutConfig(configLine.LineContents); isTimeoutConfig {
			if config.Timeout > 0 {
				// !! TODO !! Can have Query delimiter set n times
				s.setFatalMessage("Timeout config set twice", configLine.SourceLineIndex)
				return config
			}

			if err != nil {
				s.setFatalMessage(err.Error(), configLine.SourceLineIndex)
				return config
			}

			config.Timeout = timeoutValue
			continue
		}

		if isQueryDelim, queryDelimValue, err := parseQueryDelim(configLine.LineContents); isQueryDelim {
			if config.QueryDelim != nil {
				// !! TODO !! Can have Query delimiter set n times
				s.setFatalMessage("Query delimiter set twice", configLine.SourceLineIndex)
				return config
			}

			if err != nil {
				s.setFatalMessage(err.Error(), configLine.SourceLineIndex)
				return config
			}

			config.QueryDelim = &queryDelimValue
			continue
		}
	}

	return config
}
