package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	C_ARITHMETIC = iota
	C_PUSH
	C_POP
	/*
		C_LABEL
		C_GOTO
		C_IF
		C_FUNCTION
		C_RETURN
		C_CALL
	*/
)

type Parser struct {
	scanner        *bufio.Scanner
	CurrentCommand string
	NextCommand    string
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
		scanner: scanner,
	}

	p.Advance()

	return p, nil
}

func verify(argument string) error {
	log.Printf("Argument: %s\n", argument)

	if !strings.HasSuffix(argument, vmExtension) {
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
	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}

		if line == "" {
			continue
		}

		p.NextCommand = line

		return
	}
}

func (p *Parser) CommandType() int {
	switch {
	case strings.HasPrefix(p.CurrentCommand, "push"):
		return C_PUSH
	case strings.HasPrefix(p.CurrentCommand, "pop"):
		return C_POP
	case strings.HasPrefix(p.CurrentCommand, "add"):
		return C_ARITHMETIC
	default:
		return -1
	}
}

func (p *Parser) Arg1() string {
	cType := p.CommandType()
	switch cType {
	case C_ARITHMETIC:
		return p.CurrentCommand
	case C_POP, C_PUSH:
		fields := strings.Fields(p.CurrentCommand)
		if len(fields) < 2 {
			return "" // unexpected
		}
		return fields[1]
	default:
		return ""
	}
}

func (p *Parser) Arg2() int {
	cType := p.CommandType()
	switch cType {
	case C_ARITHMETIC:
		return -1 // unexpected
	case C_POP, C_PUSH:
		fields := strings.Fields(p.CurrentCommand)
		if len(fields) < 3 {
			return -2 // unexpected
		}
		arg2, err := strconv.Atoi(fields[2])
		if err != nil {
			return -3
		}
		return arg2
	default:
		return -4
	}
}

func (p *Parser) Err() error {
	// bytes reader should never return an error but including in case the implementation changes and as a best practice
	return p.scanner.Err()
}
