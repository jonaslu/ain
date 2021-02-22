package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jonaslu/ain/internal/pkg/call"
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

	// !! TODO !! Hook into SIGINT etc and cancel this context if hit
	ctx := context.Background()

	callData, fatals := parse.ParseTemplate(ctx, template)
	if len(fatals) > 0 {
		for _, fatal := range fatals {
			fmt.Println(fatal)
		}

		os.Exit(1)
	}

	curlOutput, err := call.Curl(ctx, callData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, curlOutput)

	// !! TODO !! Print errors to stderr

	// ~/.ain/ain.conf
	// ~/.ain/global.ain

	// -e execute, do not edit meld <(ain -e -h 1) <(ain -e file.ain)
	// -h history
	// -h 1 first in history
	// -i ignore global
	// -c insert global as comments
	// -p print the command, don't run it. Allows for ain test.ain > share_me.sh
	// -v verbose (print subshell results, curl command line)

	/* If body too big (like 500 characters - save it in a temp-file in .ain/
	use folders so I can wipe the folder when storing 1-10 calls */

	// default is file first and global last if collision
}
