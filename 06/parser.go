package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	A_COMMAND = iota
	C_COMMAND
	L_COMMAND
)

type parser struct {
	scanner        *bufio.Scanner
	currentCommand string
	nextCommand    string
	asmFileName    string
	hackFileName   string
	filePath       string
	data           []byte
}

func NewParser(argument string) (*parser, error) {
	if err := verify(argument); err != nil {
		return nil, err
	}

	fileName := filepath.Base(argument)
	filePath := filepath.Dir(argument)

	log.Printf("FileName: %s\n", fileName)
	log.Printf("FilePath: %s\n", filePath)

	trimmedFileName := strings.TrimSuffix(fileName, asmExtension)
	hackFileName := trimmedFileName + hackExtension

	data, err := os.ReadFile(argument)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))

	p := &parser{
		scanner:      scanner,
		asmFileName:  fileName,
		hackFileName: hackFileName,
		filePath:     filePath,
		data:         data,
	}

	p.advance()

	return p, nil
}

func verify(argument string) error {
	log.Printf("Argument: %s\n", argument)

	if !strings.HasSuffix(argument, asmExtension) {
		return fmt.Errorf("file extension must be .asm: %s", argument)
	}

	return nil
}

// hasMoreCommands will return true if there are more lines to read in the file
func (p *parser) hasMoreCommands() bool {
	return p.nextCommand != ""
}

// advance will set the current command to the next available line of the file
func (p *parser) advance() {
	p.currentCommand = p.nextCommand

	p.nextCommand = ""
	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}

		if line == "" {
			continue
		}

		p.nextCommand = line

		return
	}
}

func (p *parser) reset() {
	p.scanner = bufio.NewScanner(bytes.NewReader(p.data))
	p.currentCommand = ""
	p.nextCommand = ""
	p.advance()
}

// commandType returns the command type of the currently stored command
func (p *parser) commandType() int {
	switch {
	case strings.HasPrefix(p.currentCommand, "@"):
		return A_COMMAND
	case strings.HasPrefix(p.currentCommand, "("):
		return L_COMMAND
	default:
		return C_COMMAND
	}
}

// symbol returns the symbol or decimal value of the current command.
// Should only be called when commandType() returns an A_COMMAND or L_COMMAND
func (p *parser) symbol() string {
	var symbol string
	switch p.commandType() {
	case A_COMMAND:
		symbol = strings.TrimPrefix(p.currentCommand, "@")
	case L_COMMAND:
		symbol = strings.TrimPrefix(p.currentCommand, "(")
		symbol = strings.TrimSuffix(symbol, ")")
	}

	return symbol
}

// dest returns the dest mnemonic in the current command.
// Should only be called when commandType() returns a C_COMMAND
func (p *parser) dest() string {
	dest := ""
	if strings.Contains(p.currentCommand, "=") {
		parts := strings.Split(p.currentCommand, "=")
		if len(parts) >= 2 {
			dest = strings.TrimSpace(parts[0])
		}
	}

	return dest
}

// comp returns the comp mnemonic in the current command.
// Should only be called when commandType() returns a C_COMMAND
func (p *parser) comp() string {
	command := p.currentCommand
	if strings.Contains(command, "=") {
		parts := strings.Split(command, "=")
		if len(parts) >= 2 {
			command = strings.TrimSpace(parts[1])
		}
	}

	if strings.Contains(command, ";") {
		parts := strings.Split(command, ";")
		command = strings.TrimSpace(parts[0])
	}

	return command
}

// jump returns the jump mnemonic in the current command.
// Should only be called when commandType() returns a C_COMMAND
func (p *parser) jump() string {
	jump := ""
	if strings.Contains(p.currentCommand, ";") {
		parts := strings.Split(p.currentCommand, ";")
		if len(parts) >= 2 {
			jump = strings.TrimSpace(parts[1])
		}
	}

	return jump
}
