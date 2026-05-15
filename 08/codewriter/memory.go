package codewriter

import (
	"fmt"

	"github.com/ctjnkns/nand2tetris/08/vmtranslator/parser"
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
