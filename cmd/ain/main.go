package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/disk"
	"github.com/jonaslu/ain/internal/pkg/parse"
)

var version = "1.4.1"
var gitSha = "develop"

const bashSignalCaughtBase = 128

func printInternalErrorAndExit(err error) {
	formattedError := fmt.Errorf("Error: %v", err.Error())
	fmt.Fprintln(os.Stderr, formattedError.Error())
	os.Exit(1)
}

func checkSignalRaisedAndExit(ctx context.Context, signalRaised os.Signal) {
	if ctx.Err() == context.Canceled {
		if sigValue, ok := signalRaised.(syscall.Signal); ok {
			os.Exit(bashSignalCaughtBase + int(sigValue))
		}

		os.Exit(1)
	}
}

func main() {
	var leaveTmpFile, printCommand, showVersion, generateEmptyTemplate bool
	var envFile string

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		introMsg := `Ain is an HTTP API client. It reads template files to make the HTTP call.
These can be given on the command line or sent over a pipe.

Project home page: https://github.com/jonaslu/ain`

		fmt.Fprintf(w, "%s\n\nusage: %s [options]... <template.ain>...\n", introMsg, os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&leaveTmpFile, "l", false, "Leave any temp-files")
	flag.BoolVar(&printCommand, "p", false, "Print command to the terminal instead of executing")
	flag.StringVar(&envFile, "e", ".env", "Path to .env file")
	flag.BoolVar(&showVersion, "v", false, "Show version and exit")
	flag.BoolVar(&generateEmptyTemplate, "b", false, "Generate basic template files(s)")
	flag.Parse()

	if showVersion {
		fmt.Printf("Ain %s (%s) %s/%s\n", version, gitSha, runtime.GOOS, runtime.GOARCH)
		return
	}

	if generateEmptyTemplate {
		if err := disk.GenerateEmptyTemplates(); err != nil {
			printInternalErrorAndExit(err)
		}

		return
	}

	if err := disk.ReadEnvFile(envFile, envFile != ".env"); err != nil {
		printInternalErrorAndExit(err)
	}

	localTemplateFileNames, err := disk.GetTemplateFilenames()
	if err != nil {
		printInternalErrorAndExit(err)
	}

	if len(localTemplateFileNames) == 0 {
		// !! TODO !! This is not an internal error
		printInternalErrorAndExit(errors.New("Missing template file name(s)\n\nTry 'ain -h' for more information"))
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	var signalRaised os.Signal

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		signalRaised = <-sigs
		cancel()
	}()

	assembledCtx, backendInput, fatal, err := parse.Assemble(cancelCtx, localTemplateFileNames)
	if err != nil {
		checkSignalRaisedAndExit(assembledCtx, signalRaised)

		printInternalErrorAndExit(err)
	}

	if fatal != "" {
		// Is this valid?
		checkSignalRaisedAndExit(assembledCtx, signalRaised)

		fmt.Fprintln(os.Stderr, fatal)
		os.Exit(1)
	}

	backendInput.PrintCommand = printCommand

	call, err := call.Setup(backendInput)
	if err != nil {
		printInternalErrorAndExit(err)
	}

	if printCommand {
		// Tempfile always left when calling as string
		fmt.Fprint(os.Stdout, call.CallAsString())
		return
	}

	var errors []string
	backendInput.LeaveTempFile = leaveTmpFile
	backendOutput, err := call.CallAsCmd(assembledCtx)

	teardownErr := call.Teardown()
	if teardownErr != nil {
		errors = append(errors, teardownErr.Error())
	}

	if err != nil && assembledCtx.Err() != context.Canceled {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		errorMsg := "Error"
		if len(errors) > 1 {
			errorMsg += "s:\n"
		} else {
			errorMsg += ": "
		}

		errorMsg += strings.Join(errors, "\n") + "\n"
		fmt.Fprintln(os.Stderr, errorMsg)
	}

	if backendOutput != nil {
		// It's customary to print stderr first
		// to get the users attention on the error
		fmt.Fprint(os.Stderr, backendOutput.Stderr)
		fmt.Fprint(os.Stdout, backendOutput.Stdout)
	}

	checkSignalRaisedAndExit(assembledCtx, signalRaised)

	if assembledCtx.Err() == context.DeadlineExceeded || teardownErr != nil {
		os.Exit(1)
	}

	os.Exit(backendOutput.ExitCode)
}
