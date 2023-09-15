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
	allSections := []*Sections{}
	fatals := []string{}

	for _, filename := range filenames {
		// !! TODO !! The file-name will be displayed as test.ain! <- Remove the exclamation-mark
		// when setting the file-name2.
		rawTemplateString, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			return nil, "", err
		}

		if sections := NewSections(rawTemplateString, filename); sections.HasFatalMessages() {
			fatals = append(fatals, sections.GetFatalMessages())
		} else {
			allSections = append(allSections, sections)
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	spew.Dump(allSections)

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
}

