package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"unicode"
)

const historyFilename = ".gosh_history"

type Shell struct {
	workingDir      string
	signalChan      chan os.Signal
	historyFilepath string
	history         []string
	historyPos      int
	input           string
	lastPrinted     int
}

func NewShell() (*Shell, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	historyPath := path.Join(userDir, historyFilename)

	return &Shell{
		workingDir:      pwd,
		signalChan:      make(chan os.Signal),
		historyFilepath: historyPath,
	}, nil
}

func (s *Shell) insertChar(c byte) {
	s.input += string(c)
}

func (s *Shell) deleteChar() {
	if len(s.input) == 0 {
		return
	}
	s.input = s.input[:len(s.input)-1]
}

func (s *Shell) loadHistory() error {
	data, err := os.ReadFile(s.historyFilepath)
	if err != nil {
		return err
	}

	s.history = strings.Split(string(data), "\n")
	return nil
}

func (s *Shell) isValidChar(b byte) bool {
	if b == '\n' {
		return true
	}
	if b == '[' {
		return false
	}
	r := rune(b)
	return unicode.IsSpace(r) || unicode.IsDigit(r) || unicode.IsLetter(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
}
func (s *Shell) saveHistory() error {
	data := strings.Join(s.history, "\n")
	return os.WriteFile(s.historyFilepath, []byte(data), os.ModePerm)
}

func (s *Shell) Start(ctx context.Context) error {
	signal.Notify(s.signalChan, os.Interrupt)

	s.loadHistory()
	defer s.saveHistory()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.Prompt()
		}
	}
}

func (s *Shell) previousCommand() string {
	idx := len(s.history) - s.historyPos - 1
	if idx >= 0 && idx < len(s.history) {
		s.historyPos += 1
		cmd := s.history[idx]
		return cmd
	}
	fmt.Print("\a")
	return s.input
}
func (s *Shell) nextCommand() string {
	idx := len(s.history) - s.historyPos + 1
	if idx >= 0 && idx < len(s.history) {
		s.historyPos -= 1
		cmd := s.history[idx]
		return cmd
	}
	fmt.Print("\a")
	return s.input
}

func (s *Shell) readInput() (string, error) {
	scanner := bufio.NewReader(os.Stdin)

	s.input = ""
	s.historyPos = 0

	var prev byte
	for {
		s.printPrompt()

		b, err := scanner.ReadByte()
		if err != nil {
			fmt.Println("error: ", err.Error())
			break
		}

		if prev == '[' {
			switch b {
			case 'A':
				// up arrow
				s.input = s.previousCommand()
				prev = 0
				continue
			case 'B':
				// down arrow
				s.input = s.nextCommand()
				prev = 0
				continue
			case 'D':
				// left arrow
				prev = 0
				continue
			case 'C':
				// right arrow
				prev = 0
				continue
			default:
				s.insertChar(prev)
				if s.isValidChar(b) {
					s.insertChar(b)
				}
				prev = b
				continue
			}
		}

		if b == '[' {
			prev = b
			continue
		}

		// backspace
		if b == 127 {
			s.deleteChar()
		} else if s.isValidChar(b) {
			s.insertChar(b)
		}

		// enter hit
		if b == '\n' {
			break
		}

		prev = b
	}

	s.printPrompt()

	trimmedInput := strings.TrimSpace(string(s.input))
	return trimmedInput, nil
}

func (s *Shell) printPrompt() {
	if s.lastPrinted > 0 {
		fmt.Printf("\033[2K\r")
	}
	fmt.Printf("gosh > $ %s", s.input)
	s.lastPrinted = 1
}

func (s *Shell) changeDir(dir string) error {
	if !path.IsAbs(dir) {
		dir = path.Join(s.workingDir, dir)
	}

	fmt.Println("changing directory to ", dir)
	_, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	s.workingDir = dir
	return nil

}

func (s *Shell) parseCommand(input string) *exec.Cmd {
	fields := strings.Fields(input)

	commandName := fields[0]
	args := fields[1:]

	return exec.Command(commandName, args...)
}

func (s *Shell) executeCommand(cmd *exec.Cmd) error {
	err := cmd.Start()
	if err != nil {
		return err
	}

	for {
		select {
		case sig := <-s.signalChan:
			cmd.Process.Signal(sig)
			break
		default:
			return cmd.Wait()
		}
	}
}

func (s *Shell) handlePipeCommands(input string) error {
	inputs := strings.Split(input, "|")

	var commands []*exec.Cmd
	for _, input := range inputs {
		commands = append(commands, s.parseCommand(input))
	}

	for i, cmd := range commands {
		buf := &bytes.Buffer{}

		if i == len(commands)-1 {
			cmd.Stdout = os.Stdout
		} else {
			cmd.Stdout = buf
		}

		err := cmd.Run()
		if err != nil {
			return err
		}
		if i+1 < len(commands) {
			commands[i+1].Stdin = buf
		}
	}

	return nil
}

func (s *Shell) addToHistory(input string) {
	s.history = append(s.history, input)
}

func (s *Shell) Prompt() {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	input, err := s.readInput()
	if err != nil {
		fmt.Println("error reading input: ", err)
		return
	}

	// don't update history with empty input, history command, and prompts starting with a space
	if input != "" && input != "history" && input[0] != ' ' {
		defer s.addToHistory(input)
	}

	// support pipes
	if strings.Contains(input, "|") {
		s.handlePipeCommands(input)
		return
	}

	// parse the input
	fields := strings.Fields(input)

	if len(fields) == 0 {
		return
	}

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
	case "history":
		fmt.Println(strings.Join(s.history, "\n"))
		return
	case "exit":
		os.Exit(0)
	}

	// external commands
	_, err = exec.LookPath(commandName)
	if err != nil {
		fmt.Println("gosh: command not found: ", commandName)
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
