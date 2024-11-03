package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/jonaslu/ain/internal/app/ain"
	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/disk"
	"github.com/jonaslu/ain/internal/pkg/parse"
)

var version = "1.5.0"
var gitSha = "develop"

const bashSignalCaughtBase = 128

func printErrorAndExit(err error) {
	formattedError := fmt.Sprintf("Error: %s", err.Error())
	fmt.Fprintln(os.Stderr, formattedError)
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
	cmdParams := ain.NewCmdParams()

	if cmdParams.ShowVersion {
		fmt.Printf("Ain %s (%s) %s/%s\n", version, gitSha, runtime.GOOS, runtime.GOARCH)
		return
	}

	if err := cmdParams.SetEnvVarsAndFilenames(); err != nil {
		printErrorAndExit(err)
	}

	if cmdParams.GenerateEmptyTemplate {
		if err := disk.GenerateEmptyTemplates(cmdParams.TemplateFileNames); err != nil {
			printErrorAndExit(err)
		}

		return
	}

	for _, envVars := range cmdParams.EnvVars {
		varName := envVars[0]
		value := envVars[1]
		os.Setenv(varName, value)
	}

	if err := disk.ReadEnvFile(cmdParams.EnvFile, cmdParams.EnvFile != ".env"); err != nil {
		printErrorAndExit(err)
	}

	localTemplateFileNames, err := disk.GetTemplateFilenames(cmdParams.TemplateFileNames)
	if err != nil {
		printErrorAndExit(err)
	}

	if len(localTemplateFileNames) == 0 {
		printErrorAndExit(fmt.Errorf("missing template file name(s)\n\nTry 'ain -h' for more information"))
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

		printErrorAndExit(err)
	}

	if fatal != "" {
		// Is this valid?
		checkSignalRaisedAndExit(assembledCtx, signalRaised)

		fmt.Fprintln(os.Stderr, fatal)
		os.Exit(1)
	}

	backendInput.PrintCommand = cmdParams.PrintCommand

	call, err := call.Setup(backendInput)
	if err != nil {
		printErrorAndExit(err)
	}

	if cmdParams.PrintCommand {
		// Tempfile always left when calling as string
		fmt.Fprint(os.Stdout, call.CallAsString())
		return
	}

	var errors []string
	backendInput.LeaveTempFile = cmdParams.LeaveTmpFile
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
