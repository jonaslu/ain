package ain

import (
	"fmt"
	"os"
	"strings"
)

func printUsage(appName string, flags []flag) {
	w := os.Stderr

	introMsg := `Ain is an HTTP API client. It reads template files to make the HTTP call.
These can be given on the command line or sent over a pipe.

Project home page: https://github.com/jonaslu/ain`

	fmt.Fprintf(w, "%s\n\nusage: %s [OPTIONS] <template.ain> [--vars VAR=VALUE ...] \n", introMsg, appName)
	fmt.Fprintf(w, "\nOPTIONS:\n")
	for _, f := range flags {
		fmt.Fprintf(w, "  %-22s %s\n", f.flagName, f.usage)
	}

	fmt.Fprintf(w, "\nARGUMENTS:\n")
	fmt.Fprintf(w, "  <template.ain>[!]       One or more template files to process. Required\n")
	fmt.Fprintf(w, "  --vars VAR=VALUE [...]  Values for environment variables, set after <template.ain> file(s)\n")
}

type flagConsumer func([]string) (found bool, restArgs []string, error error)

type flag struct {
	flagName     string
	usage        string
	flagConsumer flagConsumer
}

func makeBoolConsumer(flagName string, val *bool) flagConsumer {
	return func(args []string) (bool, []string, error) {
		if args[0] == flagName {
			*val = true
			return true, args[1:], nil
		}

		return false, args, nil
	}
}

func makeStringConsumer(flagName string, val *string) flagConsumer {
	return func(args []string) (bool, []string, error) {
		if args[0] == flagName {
			if len(args) < 2 {
				return false, args, fmt.Errorf("flag %s requires an argument", flagName)
			}

			*val = args[1]

			return true, args[2:], nil
		}

		return false, args, nil
	}
}

func makeRedefinedGuardConsumer(name string, flagConsumer flagConsumer) flagConsumer {
	consumed := false
	return func(args []string) (bool, []string, error) {
		consumerConsumed, restArgs, err := flagConsumer(args)

		if consumerConsumed && consumed {
			return false, restArgs, fmt.Errorf("flag %s passed twice", name)
		}

		consumed = consumerConsumed

		return consumed, restArgs, err
	}
}

func makeFlag(flagName, usage string, flagConsumer flagConsumer) flag {
	return flag{
		flagName:     flagName,
		usage:        usage,
		flagConsumer: flagConsumer,
	}
}

func makeBoolFlag(flagName, usage string, val *bool) flag {
	return makeFlag(flagName, usage, makeRedefinedGuardConsumer(flagName, makeBoolConsumer(flagName, val)))
}

func makeStringFlag(flagName, usage string, val *string) flag {
	return makeFlag(flagName, usage, makeRedefinedGuardConsumer(flagName, makeStringConsumer(flagName, val)))
}

func NewCmdParams() *CmdParams {
	var leaveTmpFile, printCommand, showVersion, generateEmptyTemplate, showHelp bool
	envFile := ".env"

	flags := []flag{}

	appName := os.Args[0]
	restArgs := os.Args[1:]

	flags = append(flags, makeBoolFlag("-p", "Print command to the terminal instead of executing", &printCommand))
	flags = append(flags, makeStringFlag("-e", "Path to .env file", &envFile))
	flags = append(flags, makeBoolFlag("-l", "Leave any body-files", &leaveTmpFile))
	flags = append(flags, makeBoolFlag("-b", "Generate basic template files(s)", &generateEmptyTemplate))
	flags = append(flags, makeBoolFlag("-v", "Show version and exit", &showVersion))
	flags = append(flags, makeBoolFlag("-h", "Show help and exit", &showHelp))

	for {
		if len(restArgs) == 0 {
			break
		}

		arg := restArgs[0]

		if arg == "--vars" {
			break
		}

		if strings.HasPrefix(arg, "-") {
			flagFound := false

			for _, flag := range flags {
				consumed, consumedRestArgs, err := flag.flagConsumer(restArgs)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)
					os.Exit(1)
				}

				restArgs = consumedRestArgs
				if consumed {
					flagFound = true
					break
				}
			}

			if !flagFound {
				fmt.Fprintf(os.Stderr, "%s: unknown flag: %s\n", appName, arg)
				os.Exit(1)
			}

			continue
		}

		// No more flags
		break
	}

	if showHelp {
		printUsage(appName, flags)
		os.Exit(0)
	}

	return &CmdParams{
		restArgs:              restArgs,
		LeaveTmpFile:          leaveTmpFile,
		PrintCommand:          printCommand,
		ShowVersion:           showVersion,
		GenerateEmptyTemplate: generateEmptyTemplate,
		EnvFile:               envFile,
	}
}

func (c *CmdParams) SetEnvVarsAndFilenames() error {
	collectVars := false
	vars := []string{}

	for _, arg := range c.restArgs {
		if arg == "--vars" {
			collectVars = true
			continue
		}

		if collectVars {
			vars = append(vars, arg)
		} else {
			c.TemplateFileNames = append(c.TemplateFileNames, arg)
		}
	}

	for _, v := range vars {
		varName, value, found := strings.Cut(v, "=")
		if !found {
			return fmt.Errorf("invalid environment variable format, (missing =<value>): %s", v)
		}
		c.EnvVars = append(c.EnvVars, []string{varName, value})
	}

	if collectVars && len(c.EnvVars) == 0 {
		return fmt.Errorf("--vars passed but no environment variables arguments found")
	}

	return nil
}

type CmdParams struct {
	restArgs []string

	LeaveTmpFile          bool
	PrintCommand          bool
	ShowVersion           bool
	GenerateEmptyTemplate bool
	EnvFile               string
	EnvVars               [][]string
	TemplateFileNames     []string
}
