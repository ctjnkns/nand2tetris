package codewriter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctjnkns/nand2tetris/08/vmtranslator/parser"
)

const (
	asmExtension = ".asm"
)

var segmentDict = map[string]string{
	"constant": "",
	"local":    "LCL",
	"argument": "ARG",
	"this":     "THIS",
	"that":     "THAT",
}

var directDict = map[string]string{
	"temp": "5",
}

type CodeWriter struct {
	file         *os.File
	writer       *bufio.Writer
	asmFile      string
	staticPrefix string
	labelNum     int
}

func NewCodeWriter(argument string, initSPManually bool) (*CodeWriter, error) {
	fileName := filepath.Base(argument)
	filePath := filepath.Dir(argument)

	trimmedFileName := strings.TrimSuffix(fileName, parser.VMExtension)
	asmFileName := trimmedFileName + asmExtension

	asmFile := filepath.Join(filePath, asmFileName)
	f, err := os.Create(asmFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create .asm file: %v", err)
	}

	w := bufio.NewWriter(f)

	cw := &CodeWriter{
		file:         f,
		writer:       w,
		asmFile:      asmFile,
		staticPrefix: trimmedFileName,
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
	case parser.C_PUSH:
		switch segment {
		case "constant":
			return cw.pushConstant(index)
		case "temp":
			return cw.pushDirect(segment, index)
		case "static":
			return cw.pushStatic(segment, index)
		case "pointer":
			return cw.pushPointer(segment, index)
		default:
			return cw.pushSegment(segment, index)
		}
	case parser.C_POP:
		switch segment {
		case "temp":
			return cw.popDirect(segment, index)
		case "static":
			return cw.popStatic(segment, index)
		case "pointer":
			return cw.popPointer(segment, index)
		default:
			return cw.popSegment(segment, index)
		}
	default:
		return fmt.Errorf("received invalid PushPop command: %d", command)
	}
}

func (cw *CodeWriter) pushStatic(segment string, index int) error {
	symbol := fmt.Sprintf("%s.%d", cw.staticPrefix, index)
	lines := []string{
		fmt.Sprintf("// push %s %d", segment, index),
		fmt.Sprintf("@%s", symbol),
		"D=M", // Save the value in the static var to D

		"@SP",
		"A=M",
		"M=D", // save the locally saved value to the stack
		"@SP",
		"M=M+1", // bump the stack pointer
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) pushPointer(segment string, index int) error {
	var virtSeg string
	switch index {
	case 0:
		virtSeg = "THIS"
	case 1:
		virtSeg = "THAT"
	default:
		return fmt.Errorf("received unsupported pointer index: %d", index)
	}

	lines := []string{
		fmt.Sprintf("// push %s %d", segment, index),
		fmt.Sprintf("@%s", virtSeg),
		"D=M", // Save the value in the base location to D

		"@SP",
		"A=M",
		"M=D", // save the locally saved value to the stack
		"@SP",
		"M=M+1", // bump the stack pointer
	}

	return cw.writeLines(lines)
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

func (cw *CodeWriter) pushDirect(segment string, index int) error {
	base, ok := directDict[segment]
	if !ok {
		return fmt.Errorf("push: base address not found in map: %s", segment)
	}

	lines := []string{
		fmt.Sprintf("// push %s %d", segment, index),
		fmt.Sprintf("@%d", index),
		"D=A",
		fmt.Sprintf("@%s", base),
		"A=D+A",
		"D=M", // Save the value in the full virt seg location to D

		"@SP",
		"A=M",
		"M=D", // save the locally saved value to the stack
		"@SP",
		"M=M+1", // bump the stack pointer
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) pushSegment(segment string, index int) error {
	virtSeg, ok := segmentDict[segment]
	if !ok {
		return fmt.Errorf("push: virtual segment not found in map: %s", segment)
	}

	lines := []string{
		fmt.Sprintf("// push %s %d", segment, index),
		fmt.Sprintf("@%d", index),
		"D=A",
		fmt.Sprintf("@%s", virtSeg),
		"A=D+M",
		"D=M", // Save the value in the full virt seg location to D

		"@SP",
		"A=M",
		"M=D", // save the locally saved value to the stack
		"@SP",
		"M=M+1", // bump the stack pointer
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) popStatic(segment string, index int) error {
	symbol := fmt.Sprintf("%s.%d", cw.staticPrefix, index)
	lines := []string{
		fmt.Sprintf("// pop %s %d", segment, index),
		"@SP",   // Get the SP pointer
		"M=M-1", // decrement SP pointer to get to the active location
		"A=M",
		"D=M", // save the stack value to D

		fmt.Sprintf("@%s", symbol), // get the virt seg register
		"M=D",                      // Set static symbol to the value popped from the stack
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) popPointer(segment string, index int) error {
	var virtSeg string
	switch index {
	case 0:
		virtSeg = "THIS"
	case 1:
		virtSeg = "THAT"
	default:
		return fmt.Errorf("received unsupported pointer index: %d", index)
	}

	lines := []string{
		fmt.Sprintf("// pop %s %d", segment, index),

		"@SP",   // Get the SP pointer
		"M=M-1", // decrement SP pointer to get to the active location
		"A=M",
		"D=M", // save the stack value to D

		fmt.Sprintf("@%s", virtSeg), // get the virt seg register
		"M=D",                       // set THIS/THAT
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) popDirect(segment string, index int) error {
	base, ok := directDict[segment]
	if !ok {
		return fmt.Errorf("pop: base address not found in map: %s", segment)
	}

	lines := []string{
		fmt.Sprintf("// pop %s %d", segment, index),
		fmt.Sprintf("@%d", index), // Save the index to D
		"D=A",
		fmt.Sprintf("@%s", base), // get the base address
		"A=D+A",                  // Set A to the base + index
		"D=A",                    // save to D
		"@R13",
		"M=D", // store full virt seg register address in R13 scratch location since we need D for popping from stack

		"@SP",   // Get the SP pointer
		"M=M-1", // decrement SP pointer to get to the active location
		"A=M",
		"D=M", // save the stack value to D

		"@R13", // Get the virt seg address again
		"A=M",  // Set A to RAM[R13] which was computed earlier
		"M=D",  // Set virt seg Index to the value popped from the stack
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) popSegment(segment string, index int) error {
	virtSeg, ok := segmentDict[segment]
	if !ok {
		return fmt.Errorf("pop: virtual segment not found in map: %s", segment)
	}

	lines := []string{
		fmt.Sprintf("// pop %s %d", segment, index),
		fmt.Sprintf("@%d", index), // Save the index to D
		"D=A",
		fmt.Sprintf("@%s", virtSeg), // get the virt seg register
		"A=D+M",                     // Set A to the virt seg base + index
		"D=A",                       // save to D
		"@R13",
		"M=D", // store full virt seg register address in R13 scratch location since we need D for popping from stack

		"@SP",   // Get the SP pointer
		"M=M-1", // decrement SP pointer to get to the active location
		"A=M",
		"D=M", // save the stack value to D

		"@R13", // Get the virt seg address again
		"A=M",  // Set A to RAM[R13] which was computed earlier
		"M=D",  // Set virt seg Index to the value popped from the stack
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) WriteLabel(label string) error {
	lines := []string{
		fmt.Sprintf("// label %s", label),
		fmt.Sprintf("(%s)", label),
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) WriteIf(label string) error {
	lines := []string{
		fmt.Sprintf("// if-goto %s", label),
		"@SP",   // Get the SP pointer
		"M=M-1", // decrement SP pointer to get to the active location
		"A=M",
		"D=M", // save the stack value to D

		fmt.Sprintf("@%s", label),
		"D;JNE",
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
