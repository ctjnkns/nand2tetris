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
	file     *os.File
	writer   *bufio.Writer
	asmFile  string
	labelNum int
}

func NewCodeWriter(argument string, initSPManually bool) (*CodeWriter, error) {
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

	if initSPManually {
		err := cw.initSP()
		if err != nil {
			return nil, err
		}
	}

	return cw, nil
}

func (cw *CodeWriter) initSP() error {
	lines := []string{
		"// init SP",
		"@256",
		"D=A",
		"@SP",
		"M=D",
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) WriteArithmetic(command string) error {
	switch command {
	case "add":
		return cw.add()
	case "sub":
		return cw.sub()
	case "and":
		return cw.and()
	case "or":
		return cw.or()
	case "eq":
		return cw.eq()
	case "lt":
		return cw.lt()
	case "gt":
		return cw.gt()
	case "neg":
		return cw.neg()
	case "not":
		return cw.not()
	default:
		return fmt.Errorf("received invalid arithmetic operation: %s", command)
	}
}

func (cw *CodeWriter) add() error {
	return cw.xyOp("add", "D+M")
}

func (cw *CodeWriter) sub() error {
	return cw.xyOp("sub", "M-D")
}

func (cw *CodeWriter) and() error {
	return cw.xyOp("and", "D&M")
}

func (cw *CodeWriter) or() error {
	return cw.xyOp("or", "D|M")
}

func (cw *CodeWriter) xyOp(name, op string) error {
	lines := []string{
		fmt.Sprintf("// %s", name),
		"@SP",                   // pop y
		"M=M-1",                 // sp--
		"A=M",                   // Set A to RAM[0]
		"D=M",                   // Save y to D register
		"@SP",                   // pop x
		"M=M-1",                 // sp--
		"A=M",                   // Set A to RAM[0]
		fmt.Sprintf("M=%s", op), // x & y
		"@SP",                   // bump SP
		"M=M+1",
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) eq() error {
	return cw.xyCompOp("eq", "JEQ")
}

func (cw *CodeWriter) lt() error {
	return cw.xyCompOp("lt", "JLT")
}

func (cw *CodeWriter) gt() error {
	return cw.xyCompOp("gt", "JGT")
}

func (cw *CodeWriter) xyCompOp(name, op string) error {
	cw.labelNum++ // increment the labelNum so that when this function is called there are no label collisions, each label within these lines must be unique
	lines := []string{
		fmt.Sprintf("// %s", name),
		"@SP",   // pop y
		"M=M-1", // sp--
		"A=M",   // set A to RAM[0]
		"D=M",   // save y to D register
		"@SP",   // pop x
		"M=M-1", // sp--
		"A=M",   // set a to RAM[0]
		"D=M-D",
		fmt.Sprintf("@%s_%d", name, cw.labelNum), // set A to the label location for if they're equal
		fmt.Sprintf("D;%s", op),                  // jump if M-D is < 0
		"@SP",
		"A=M", // Set A to RAM[0]
		"M=0", // save 0 to the SP pointer location
		fmt.Sprintf("@END_%d", cw.labelNum),
		"0;JMP",
		fmt.Sprintf("(%s_%d)", name, cw.labelNum), // equal condition
		"@SP", // push -1 to the stack
		"A=M", // set A to RAM[0]
		"M=-1",
		fmt.Sprintf("(END_%d)", cw.labelNum),
		"@SP", // bump SP
		"M=M+1",
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) neg() error {
	return cw.yOp("neg", "-M")
}

func (cw *CodeWriter) not() error {
	return cw.yOp("not", "!M")
}

func (cw *CodeWriter) yOp(name, op string) error {
	lines := []string{
		fmt.Sprintf("// %s", name),
		"@SP",
		"M=M-1", // pop y
		"A=M",
		fmt.Sprintf("M=%s", op),
		"@SP", // bump sp
		"M=M+1",
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
		fmt.Sprintf("// push constant %d", index),
		fmt.Sprintf("@%d", index),
		"D=A", // Set D register to the const value
		"@SP",
		"A=M",
		"M=D",

		// increment the SP pointer
		"@SP",
		"M=M+1",
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) writeLines(lines []string) error {
	for _, line := range lines {
		_, err := cw.writer.WriteString(line)
		if err != nil {
			return err
		}
		err = cw.writer.WriteByte('\n')
		if err != nil {
			return err
		}
	}

	return nil
}

func (cw *CodeWriter) Close() {
	err := cw.writer.Flush()
	if err != nil {
		log.Printf("failed to flush asm writer: %s", err)
	} else {
		log.Printf("assembly code written to: %s", cw.asmFile)
	}

	err = cw.file.Close()
	if err != nil {
		log.Printf("failed to close asm file: %v", err)
	}
}
