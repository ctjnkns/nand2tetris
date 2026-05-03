package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	asmExtension  = ".asm"
	hackExtension = ".hack"
)

type assembler struct {
	parser      *parser
	symbolTable symbolTable
}

func NewAssembler(argument string) (*assembler, error) {
	p, err := NewParser(argument)
	if err != nil {
		return nil, err
	}

	return &assembler{
		parser:      p,
		symbolTable: NewSymbolTable(),
	}, nil
}

func (a *assembler) processACommand() string {
	symbol := a.parser.symbol()
	n, err := strconv.Atoi(symbol)
	if err != nil {
		if !a.symbolTable.contains(symbol) {
			a.symbolTable.addEntry(symbol, a.symbolTable.nextSymbolAddress)
			a.symbolTable.nextSymbolAddress++
		}

		n = a.symbolTable.getAddress(symbol)
	}

	return fmt.Sprintf("0%015b", n)
}

func (a *assembler) processCCommand() string {
	comp := comp(a.parser.comp())
	dest := dest(a.parser.dest())
	jump := jump(a.parser.jump())

	return "111" + comp + dest + jump
}

func (a *assembler) firstPass() {
	romAddr := 0
	for a.parser.hasMoreCommands() {
		a.parser.advance()

		switch a.parser.commandType() {
		case L_COMMAND:
			symbol := a.parser.symbol()
			if !a.symbolTable.contains(symbol) {
				a.symbolTable.addEntry(symbol, romAddr)
			}
		case A_COMMAND, C_COMMAND:
			romAddr++
		}
	}

	a.parser.reset()
}

func (a *assembler) secondPass() error {
	var sb strings.Builder
	first := true
	for a.parser.hasMoreCommands() {
		a.parser.advance()

		bLine := ""
		switch a.parser.commandType() {
		case A_COMMAND:
			bLine = a.processACommand()
		case C_COMMAND:
			bLine = a.processCCommand()
		}

		if bLine == "" {
			continue
		}

		if !first {
			sb.WriteByte('\n')
		}
		sb.WriteString(bLine)
		first = false
	}

	// bytes reader should never return an error but including in case the implementation changes and as a best practice
	if err := a.parser.scanner.Err(); err != nil {
		return fmt.Errorf("scan error on exit: %v", err)
	}

	hackFile := filepath.Join(a.parser.filePath, a.parser.hackFileName)

	err := os.WriteFile(hackFile, []byte(sb.String()), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write hack file: %v", err)
	}

	log.Printf("Successfully wrote file: %s", hackFile)

	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// Hack Assembler
func run(args []string) error {
	if len(args) < 2 {
		return errors.New("must provide .asm file as an argument")
	}

	argument := args[1]

	// initialize
	assembler, err := NewAssembler(argument)
	if err != nil {
		return fmt.Errorf("failed to initialize assembler: %v", err)
	}

	assembler.firstPass()

	return assembler.secondPass()
}
