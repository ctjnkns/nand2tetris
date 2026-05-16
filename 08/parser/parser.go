package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	C_ARITHMETIC = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

const (
	VMExtension = ".vm"
)

var arithmeticCommands = []string{
	"add",
	"sub",
	"neg",
	"eq",
	"gt",
	"lt",
	"and",
	"or",
	"not",
}

type Parser struct {
	Scanner        *bufio.Scanner
	CurrentCommand string
	NextCommand    string
	FileName       string
}

func NewParser(argument string) (*Parser, error) {
	if err := verify(argument); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(argument)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))

	p := &Parser{
		Scanner:  scanner,
		FileName: strings.TrimSuffix(filepath.Base(argument), VMExtension),
	}

	p.Advance()

	return p, nil
}

func verify(argument string) error {
	log.Printf("Argument: %s\n", argument)

	if !strings.HasSuffix(argument, VMExtension) {
		return fmt.Errorf("file extension must be .vm: %s", argument)
	}

	return nil
}

func (p *Parser) HasMoreLines() bool {
	return p.NextCommand != ""
}

func (p *Parser) Advance() {
	p.CurrentCommand = p.NextCommand

	p.NextCommand = ""
	for p.Scanner.Scan() {
		line := strings.TrimSpace(p.Scanner.Text())
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}

		if line == "" {
			continue
		}

		p.NextCommand = line

		return
	}
}

func (p *Parser) CommandType() (int, error) {
	switch {
	case strings.HasPrefix(p.CurrentCommand, "push"):
		return C_PUSH, nil
	case strings.HasPrefix(p.CurrentCommand, "pop"):
		return C_POP, nil
	case strings.HasPrefix(p.CurrentCommand, "label"):
		return C_LABEL, nil
	case strings.HasPrefix(p.CurrentCommand, "if-goto"):
		return C_IF, nil
	case strings.HasPrefix(p.CurrentCommand, "goto"):
		return C_GOTO, nil
	case strings.HasPrefix(p.CurrentCommand, "call"):
		return C_CALL, nil
	case strings.HasPrefix(p.CurrentCommand, "function"):
		return C_FUNCTION, nil
	case strings.HasPrefix(p.CurrentCommand, "return"):
		return C_RETURN, nil
	case slices.Contains(arithmeticCommands, p.CurrentCommand):
		return C_ARITHMETIC, nil
	default:
		return 0, fmt.Errorf("unknown command type: %s", p.CurrentCommand)
	}
}

func (p *Parser) Arg1() (string, error) {
	cType, err := p.CommandType()
	if err != nil {
		return "", err
	}

	switch cType {
	case C_ARITHMETIC:
		return p.CurrentCommand, nil
	case C_POP, C_PUSH, C_LABEL, C_IF, C_GOTO, C_CALL, C_FUNCTION:
		fields := strings.Fields(p.CurrentCommand)
		if len(fields) < 2 {
			return "", fmt.Errorf("arg 1: invalid vm command: %s", p.CurrentCommand) // unexpected
		}
		return fields[1], nil
	default:
		return "", fmt.Errorf("arg 1: unknown command type: %d", cType)
	}
}

func (p *Parser) Arg2() (int, error) {
	cType, err := p.CommandType()
	if err != nil {
		return 0, err
	}

	switch cType {
	case C_POP, C_PUSH, C_CALL, C_FUNCTION:
		fields := strings.Fields(p.CurrentCommand)
		if len(fields) < 3 {
			return 0, errors.New("index must be provided for pop and push commands")
		}
		arg2, err := strconv.Atoi(fields[2])
		if err != nil {
			return 0, fmt.Errorf("index must be an int, received: %s", fields[2])
		}
		return arg2, nil
	default:
		return 0, fmt.Errorf("unrecognized command type provided: %d", cType)
	}
}

func (p *Parser) Err() error {
	// bytes reader should never return an error but including in case the implementation changes and as a best practice
	return p.Scanner.Err()
}
