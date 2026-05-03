package main

import (
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
	parser *parser
}

func NewAssembler(argument string) (*assembler, error) {
	p, err := NewParser(argument)
	if err != nil {
		return nil, err
	}

	return &assembler{
		parser: p,
	}, nil
}

func (a *assembler) processACommand() string {
	symbol := a.parser.symbol()
	n, err := strconv.Atoi(symbol)
	if err != nil {
		log.Fatalf("non-numeric symbol %q: %v:", symbol, err)
	}

	return fmt.Sprintf("0%015b", n)
}

func (a *assembler) processCCommand() string {
	comp := comp(a.parser.comp())
	dest := dest(a.parser.dest())
	jump := jump(a.parser.jump())

	return "111" + comp + dest + jump
}

// Hack Assembler
func main() {
	if len(os.Args) < 2 {
		log.Fatal("must provide .asm file as an argument")
	}

	argument := os.Args[1]

	assembler, err := NewAssembler(argument)
	if err != nil {
		log.Fatalf("failed to iniialize assember: %v", err)
	}

	var sb strings.Builder
	for assembler.parser.hasMoreCommands() {
		assembler.parser.advance()

		bLine := ""
		switch assembler.parser.commandType() {
		case A_COMMAND:
			bLine = assembler.processACommand()
		case C_COMMAND:
			bLine = assembler.processCCommand()
		}

		if bLine == "" {
			continue
		}

		sb.WriteString(bLine)
		if assembler.parser.hasMoreCommands() {
			sb.WriteByte('\n')
		}
	}

	// bytes reader should never return an error but including in case the implmenetation changes and as a best practice
	if err = assembler.parser.scanner.Err(); err != nil {
		log.Fatalf("Scan error on exit: %v", err)
	}

	hackFile := filepath.Join(assembler.parser.filePath, assembler.parser.hackFileName)

	err = os.WriteFile(hackFile, []byte(sb.String()), 0644)
	if err != nil {
		log.Fatalf("failed to write hack file: %v", err)
	}

	log.Printf("Successfully wrote file: %s", hackFile)
}
