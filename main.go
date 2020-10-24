package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func main() {
	var fileName []byte

	fi, err := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		fileName, _ = ioutil.ReadAll(os.Stdin)
	} else {
		fmt.Println("Terminal")
	}

	cmd := exec.Command("vim")
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		log.Fatalf("can't open /dev/tty: %s", err)
	}

	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(fileName))
}
