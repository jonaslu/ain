package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const testBinaryPath = "./ain_test"

type testDirectives struct {
	Env       []string
	Args      []string
	AfterArgs []string `yaml:"afterArgs"`
	Stderr    string
	Stdout    string
	ExitCode  int
}

func addBarsBeforeNewlines(s string) string {
	bars := ""
	for _, c := range s {
		if c == '\n' {
			bars += "|"
		}
		bars += string(c)
	}

	return bars + "|"
}

func buildGoBinary() error {
	args := []string{"build"}
	if os.Getenv("E2EGOCOVERDIR") != "" {
		args = append(args, "-cover")
	}

	args = append(args, "-o", testBinaryPath, "../../cmd/ain/main.go")

	cmd := exec.Command("go", args...)

	err := cmd.Run()
	if err != nil {
		return errors.New("could not build binary")
	}

	return nil
}

func runTest(filename string, templateContents []byte) error {
	lines := strings.Split(string(templateContents), "\n")
	idx := len(lines)

	directives := []string{}

	for idx >= 0 {
		idx--

		line := string(lines[idx])
		if line == "" && len(directives) == 0 {
			continue
		}

		if !strings.HasPrefix(line, "# ") {
			break
		}

		trimmedPrefixLine := strings.TrimPrefix(line, "# ")

		directives = append([]string{trimmedPrefixLine}, directives...)
	}

	testDirectives := testDirectives{}
	err := yaml.Unmarshal([]byte(strings.Join(directives, "\n")), &testDirectives)

	if err != nil {
		return errors.New("Could not unmarshal yaml")
	}

	var stdout, stderr bytes.Buffer

	// !! TODO !! Get timeout from yaml or default to 1 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	totalArgs := append(testDirectives.Args, filename)
	totalArgs = append(totalArgs, testDirectives.AfterArgs...)

	cmd := exec.CommandContext(ctx, "./ain_test", totalArgs...)
	cmd.Env = testDirectives.Env
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
	if os.Getenv("E2EGOCOVERDIR") != "" {
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+os.Getenv("E2EGOCOVERDIR"))
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		return errors.New("could not start command")
	}

	err = cmd.Wait()
	if err != nil && err.(*exec.ExitError) == nil {
		return errors.New("Could not wait for command")
	}

	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("timed out")
	}

	if ctx.Err() != nil {
		return errors.New("context error")
	}

	if stderr.String() != testDirectives.Stderr {
		return fmt.Errorf("stderr %s did not match %s", addBarsBeforeNewlines(stderr.String()), addBarsBeforeNewlines(testDirectives.Stderr))
	}

	if stdout.String() != testDirectives.Stdout {
		return fmt.Errorf("stdout %s did not match %s", addBarsBeforeNewlines(stdout.String()), addBarsBeforeNewlines(testDirectives.Stdout))
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != testDirectives.ExitCode {
		return errors.New("exit code did not match")
	}

	return nil
}

func readTestFiles() (map[string][]byte, error) {
	templateFolder := "templates"
	files, err := os.ReadDir(templateFolder)

	if err != nil {
		return nil, errors.New("could not read directory")
	}

	testFiles := map[string][]byte{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := templateFolder + "/" + file.Name()
		if !strings.HasSuffix(fileName, ".ain") {
			continue
		}

		res, err := os.ReadFile(fileName)
		if err != nil {
			return nil, errors.New("could not read file")
		}

		testFiles[fileName] = res
	}

	return testFiles, nil
}

func Test_main(t *testing.T) {
	var coverage bool

	flag.BoolVar(&coverage, "coverage", false, "Enable coverage")
	flag.Parse()

	if err := buildGoBinary(); err != nil {
		t.Fatalf("Could not build binary")
		return
	}

	defer os.Remove(testBinaryPath)

	files, err := readTestFiles()
	if err != nil {
		t.Fatalf("Could not read test templates")
		return
	}

	for filename, testContents := range files {
		t.Run(filename, func(t *testing.T) {
			err := runTest(filename, testContents)
			if err != nil {
				t.Errorf("%s: %v", filename, err)
			}
		})
	}
}
