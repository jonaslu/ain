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
	AfterArgs []string `yaml:"afterargs"`
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

	for idx > 0 {
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

	testDirectives.Stdout = strings.ReplaceAll(testDirectives.Stdout, "$filename", filename)
	testDirectives.Stderr = strings.ReplaceAll(testDirectives.Stderr, "$filename", filename)

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

func runOneTest(fileName string, t *testing.T) error {
	testContents, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not read file: %s", err)
	}

	t.Run(fileName, func(t *testing.T) {
		err := runTest(fileName, testContents)
		if err != nil {
			t.Errorf("%s: %v", fileName, err)
		}
	})

	return nil
}

func readTestFiles(templateFolder string) ([]string, error) {
	files, err := os.ReadDir(templateFolder)

	if err != nil {
		return nil, errors.New("could not read directory")
	}

	testFilePaths := []string{}

	for _, file := range files {
		if file.IsDir() {
			subFolderPath := templateFolder + "/" + file.Name()
			subFolderDirs, err := readTestFiles(subFolderPath)

			if err != nil {
				return nil, errors.Join(errors.New("could not read subfolder"+subFolderPath), err)
			}

			testFilePaths = append(testFilePaths, subFolderDirs...)
			continue
		}

		fileName := templateFolder + "/" + file.Name()
		if !strings.HasSuffix(fileName, ".ain") {
			continue
		}

		testFilePaths = append(testFilePaths, fileName)

	}

	return testFilePaths, nil
}

func Test_main(t *testing.T) {
	if err := buildGoBinary(); err != nil {
		t.Fatalf("Could not build binary")
		return
	}

	defer os.Remove(testBinaryPath)

	if len(flag.Args()) > 0 {
		for _, testToRun := range flag.Args() {
			if err := runOneTest(testToRun, t); err != nil {
				t.Fatalf("Could not run test %s, error: %s", testToRun, err)
				return
			}
		}

		return
	}

	files, err := readTestFiles("templates")
	if err != nil {
		t.Fatalf("Could not read test templates")
		return
	}

	for _, fileName := range files {
		if err := runOneTest(fileName, t); err != nil {
			t.Fatalf("Could not run test %s, error: %s", fileName, err)
			return
		}
	}
}
