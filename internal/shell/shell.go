package shell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Shell struct {
	workingDir string
}

func NewShell() (*Shell, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Shell{
		workingDir: pwd}, nil
}

func (s *Shell) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.Prompt()
		}
	}
}

func (s *Shell) readInput() (string, error) {
	r := bufio.NewReader(os.Stdin)

	input, _, err := r.ReadLine()
	if err != nil {
		return "", err
	}

	trimmedInput := strings.TrimSpace(string(input))
	return trimmedInput, nil
}

func (s *Shell) printPrompt() {
	fmt.Printf("go-shell > $ ")
}

func (s *Shell) changeDir(dir string) error {
	if !path.IsAbs(dir) {
		dir = path.Join(s.workingDir, dir)
	}

	fmt.Println("changing dir to ", dir)
	_, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	s.workingDir = dir
	return nil
}

func (s *Shell) Prompt() {
	s.printPrompt()
	input, err := s.readInput()
	if err != nil {

	}

	// parse the input
	fields := strings.Fields(input)

	commandName := fields[0]
	args := fields[1:]

	cmd := exec.Command(commandName, args...)

	// built-in commands
	switch commandName {
	case "cd":
		if len(args) == 0 {
			fmt.Println("cd: requires 1 argument")
			return
		}
		err := s.changeDir(args[0])
		if err != nil {
			fmt.Println("cd: error: ", err.Error())
		}
		return
	case "pwd":
		fmt.Println(s.workingDir)
		return
	case "exit":
		os.Exit(0)
	}

	// external commands
	_, err = exec.LookPath(commandName)
	if err != nil {
		fmt.Println("go-shell: command not found: ", cmd)
		return
	}

	// set command working dir to the shell working directory
	cmd.Dir = s.workingDir

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
