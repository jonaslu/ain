package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

func checkSignalRaisedAndExit(ctx context.Context, signalRaised os.Signal) {
	if ctx.Err() == context.Canceled {
		if sigValue, ok := signalRaised.(syscall.Signal); ok {
			os.Exit(128 + int(sigValue))
		}

		os.Exit(1)
	}
}

func main() {
	var leaveTmpFile, printCommand bool
	var envFile string

	flag.BoolVar(&leaveTmpFile, "l", false, "Leave any temp-files")
	flag.BoolVar(&printCommand, "p", false, "Print command to the terminal instead of executing")
	flag.StringVar(&envFile, "e", ".env", "Path to .env file")
	flag.Parse()

	if err := disk.ReadEnvFile(envFile, envFile != ".env"); err != nil {
		printInternalErrorAndExit(err)
	}

	localTemplateFileNames, err := disk.GetTemplateFilenames()
	if err != nil {
		printInternalErrorAndExit(err)
	}

	if len(localTemplateFileNames) == 0 {
		printInternalErrorAndExit(errors.New("Missing file name\nUsage ain <template.ain> or connect it to a pipe"))
	}

	ctx, cancel := context.WithCancel(context.Background())

	var signalRaised os.Signal

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		signalRaised = <-sigs
		cancel()
	}()

	callData, fatal, err := assemble.Assemble(ctx, localTemplateFileNames)
	if err != nil {
		checkSignalRaisedAndExit(ctx, signalRaised)

		printInternalErrorAndExit(err)
	}

	if fatal != "" {
		checkSignalRaisedAndExit(ctx, signalRaised)

		fmt.Fprintln(os.Stderr, fatal)
		os.Exit(1)
	}

	backendOutput, err := call.CallBackend(ctx, callData, leaveTmpFile, printCommand)
	if err != nil {
		checkSignalRaisedAndExit(ctx, signalRaised)

		fmt.Fprint(os.Stderr, err)

		var backendErr *call.BackedErr
		if errors.As(err, &backendErr) {
			os.Exit(backendErr.ExitCode)
		}
	}

	fmt.Fprint(os.Stdout, backendOutput)
}
