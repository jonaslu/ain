package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/jonaslu/ain/internal/assemble"
	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/disk"
)

func printInternalErrorAndExit(err error) {
	formattedError := fmt.Errorf("An error occurred: %v", err.Error())
	fmt.Fprintln(os.Stderr, formattedError.Error())
	os.Exit(1)
}

func main() {
	if err := disk.ReadEnvFile(".env"); err != nil {
		printInternalErrorAndExit(err)
	}

	var execute bool

	flag.BoolVar(&execute, "execute", false, "Execute template directly, without editing")
	flag.BoolVar(&execute, "x", false, "Execute template directly, without editing")
	flag.Parse()

	localTemplateFileNames, err := disk.GetTemplateFilenames()
	if err != nil {
		printInternalErrorAndExit(err)
	}

	if len(localTemplateFileNames) == 0 {
		printInternalErrorAndExit(errors.New("Missing file name\nUsage ain <template.ain> or connect it to a pipe"))
	}

	// !! TODO !! Hook into SIGINT etc and cancel this context if hit
	ctx := context.Background()

	callData, fatal, err := assemble.Assemble(ctx, localTemplateFileNames, execute)
	if err != nil {
		printInternalErrorAndExit(err)
	}

	if fatal != "" {
		fmt.Println(fatal)
		os.Exit(1)
	}

	backendOutput, err := call.CallBackend(ctx, callData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, backendOutput)
}
