package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleError(err error) {
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

func handlePrompt() {
	fmt.Printf("go-shell > $ ")
	r := bufio.NewReader(os.Stdin)

	input, _, err := r.ReadLine()
	if err != nil {
		handleError(err)
	}

	command := strings.TrimSpace(string(input))

	cmd := exec.Command(command)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.Run()

}

func main() {
	handlePrompt()
}
