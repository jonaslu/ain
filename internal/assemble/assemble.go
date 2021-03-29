package assemble

import (
	"context"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/disk"
	"github.com/jonaslu/ain/internal/pkg/parse"
)

func mergeCallData(dest, merge *data.Parse) {
	if merge.Host != nil {
		dest.Host = merge.Host
	}

	if len(merge.Body) != 0 {
		dest.Body = merge.Body
	}

	if merge.Method != "" {
		dest.Method = merge.Method
	}

	dest.Headers = append(dest.Headers, merge.Headers...)

	if merge.Backend != "" {
		dest.Backend = merge.Backend
	}

	dest.BackendOptions = append(dest.BackendOptions, merge.BackendOptions...)

	if merge.Config.Timeout != -1 {
		dest.Config.Timeout = merge.Config.Timeout
	}
}

func validateCallData(data *data.Parse) []string {
	fatals := []string{}

	if data.Host == nil {
		fatals = append(fatals, "No mandatory [Host] section found")
	}

	if data.Backend == "" {
		fatals = append(fatals, "No mandatory [Backend] section found")
	}

	return fatals
}

func appendErrorMessages(errorMessage, filename string, fatals []string) string {
	if errorMessage != "" {
		errorMessage = errorMessage + "\n"
	}

	if filename != "" {
		errorMessage = errorMessage + `Error in file: ` + filename + "\n"
	}

	return errorMessage + strings.Join(fatals, "\n") + "\n"
}

func Assemble(ctx context.Context, filenames []string, execute bool) (*data.Parse, string, error) {
	errors := ""

	callData := &data.Parse{}
	callData.Config.Timeout = -1

	for _, filename := range filenames {
		template, err := disk.ReadTemplate(filename, execute)
		if err != nil {
			return nil, "", err
		}

		fileCallData, fatals := parse.ParseTemplate(ctx, template)
		mergeCallData(callData, fileCallData)

		if len(fatals) > 0 {
			errors = appendErrorMessages(errors, filename, fatals)
		}
	}

	if validationErrors := validateCallData(callData); len(validationErrors) > 0 {
		errors = appendErrorMessages(errors, "", validationErrors)
	}

	if errors != "" {
		return nil, errors, nil
	}

	return callData, errors, nil
}
