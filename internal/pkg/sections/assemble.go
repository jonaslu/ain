package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/disk"
)

func Assemble(ctx context.Context, filenames []string) (*data.BackendInput, string, error) {
	fatals := []string{}

	parsedTemplate := &data.ParsedTemplate{}
	parsedTemplate.Config.Timeout = data.TimeoutNotSet

	for _, filename := range filenames {
		// !! TODO !! The file-name will be displayed as test.ain! <- Remove the exclamation-mark
		// when setting the file-name2.
		rawTemplateString, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			return nil, "", err
		}

		sections := NewSections(rawTemplateString, filename)
		if sections.HasFatalMessages() {
			fatals = append(fatals, sections.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	return nil, "", nil
}

func main() {
	flag.Parse()
	filenames := flag.Args()

	backendInput, fatals, err := Assemble(context.TODO(), filenames)
	if err != nil {
		panic(err)
	}

	if fatals != "" {
		fmt.Println(fatals)
	}

	spew.Dump(backendInput)
}
