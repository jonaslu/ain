package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jonaslu/ain/internal/pkg/disk"
	"github.com/jonaslu/ain/internal/pkg/parse"
)

func printInternalErrorAndExit(err error) {
	formattedError := fmt.Errorf("An error occurred: %v", err.Error())
	fmt.Fprintln(os.Stderr, formattedError.Error())
	os.Exit(1)
}

func main() {
	gotPipe, err := disk.IsConnectedToPipe()
	if err != nil {
		printInternalErrorAndExit(err)
	}

	if !gotPipe {
		if len(os.Args) < 2 {
			printInternalErrorAndExit(errors.New("Missing file name\nUsage ain <template.ain> or connect it to a pipe"))
		}
	}

	template, err := disk.ReadTemplate()
	if err != nil {
		printInternalErrorAndExit(err)
	}

	callData, fatals := parse.ParseTemplate(template)
	fmt.Println(callData, fatals)

	// ~/.ain/ain.conf
	// ~/.ain/global.ain

	// -e execute, do not edit meld <(ain -e -h 1) <(ain -e file.ain)
	// -h history
	// -h 1 first in history
	// -i ignore global
	// -c insert global as comments

	// default is file first and global last if collision
}
