package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	vmExtension  = ".vm"
	asmExtension = ".asm"
)

type CodeWriter struct {
	file    *os.File
	writer  *bufio.Writer
	asmFile string
}

func NewCodeWriter(argument string) (*CodeWriter, error) {
	fileName := filepath.Base(argument)
	filePath := filepath.Dir(argument)

	trimmedFileName := strings.TrimSuffix(fileName, vmExtension)
	asmFileName := trimmedFileName + asmExtension

	asmFile := filepath.Join(filePath, asmFileName)
	f, err := os.Create(asmFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create .asm file: %v", err)
	}

	w := bufio.NewWriter(f)

	cw := &CodeWriter{
		file:    f,
		writer:  w,
		asmFile: asmFile,
	}

	return cw, nil
}

func (cw *CodeWriter) WriteArithmetic(command string) error {
	switch command {
	case "add":
		return cw.add()
	default:
		return fmt.Errorf("received invalid arithmetic operation: %s", command)
	}
}

func (cw *CodeWriter) add() error {
	lines := []string{
		"// add\n",
		"@SP\n",   // pop y
		"M=M-1\n", // sp--
		"A=M\n",   // Set A to RAM[0]
		"D=M\n",   // Save y to D register
		"@SP\n",   // pop x
		"M=M-1\n", // sp--
		"A=M\n",   // Set A to RAM[0]
		"M=D+M\n", // x + y
		"@SP\n",   // bump SP
		"M=M+1\n",
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) WritePushPop(command int, segment string, index int) error {
	switch command {
	case C_PUSH:
		switch segment {
		case "constant":
			return cw.pushConstant(index)
		default:
			return fmt.Errorf("received invalid push segment: %s", segment)
		}
	default:
		return fmt.Errorf("received invalid PushPop command: %d", command)
	}
}

func (cw *CodeWriter) pushConstant(index int) error {
	lines := []string{
		fmt.Sprintf("// push constant %d\n", index),
		fmt.Sprintf("@%d\n", index),
		"D=A\n", // Set D register to the const value
		"@SP\n",
		"A=M\n",
		"M=D\n",

		// increment the SP pointer
		"@SP\n",
		"M=M+1\n",
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) writeLines(lines []string) error {
	for _, line := range lines {
		_, err := cw.writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cw *CodeWriter) Close() {
	err := cw.writer.Flush()
	if err != nil {
		log.Printf("failed to flush asm writer: %s\n", err)
	} else {
		log.Printf("assembly code written to: %s", cw.asmFile)
	}

	err = cw.file.Close()
	if err != nil {
		log.Printf("failed to close asm file: %v\n", err)
	}
}
