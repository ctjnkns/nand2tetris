package codewriter

import "fmt"

func (cw *CodeWriter) WriteCall(functionName string, numArgs int) error {
	caller := cw.currentFunction
	if caller == "" {
		caller = "Bootstrap"
	}

	returnLabel := fmt.Sprintf("%s$ret.%d", caller, cw.callCount)
	cw.callCount++

	lines := []string{
		fmt.Sprintf("// call %s %d", functionName, numArgs),
		fmt.Sprintf("@%s", returnLabel), // push the return label to the stack
		"D=A",
		"@SP",
		"A=M",
		"M=D",
		"@SP", // bump SP
		"M=M+1",

		"@LCL", // push LCL to stack
		"D=M",
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",

		"@ARG", // push ARG to stack
		"D=M",
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",

		"@THIS", // push THIS to stack
		"D=M",
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",

		"@THAT", // push THIS to stack
		"D=M",
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",

		"@SP", // reposition arg
		"D=M",
		"@5",
		"D=D-A", // SP - 5
		fmt.Sprintf("@%d", numArgs),
		"D=D-A", // SP - 5 - numArgs
		"@ARG",
		"M=D",

		"@SP", // reposition lcl
		"D=M",
		"@LCL",
		"M=D",

		fmt.Sprintf("@%s", functionName),
		"0;JMP", // goto function

		fmt.Sprintf("(%s)", returnLabel), // declare the returnaddress label location
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) WriteFunction(functionName string, nVars int) error {
	cw.currentFunction = functionName
	lines := []string{
		fmt.Sprintf("// function %s %d", functionName, nVars),
		fmt.Sprintf("(%s)", functionName),
	}

	pushLocalLines := []string{
		"@SP", // init the number of local vars to 0
		"A=M",
		"M=0",
		"@SP",
		"M=M+1",
	}
	for range nVars {
		lines = append(lines, pushLocalLines...)
	}

	return cw.writeLines(lines)
}

func (cw *CodeWriter) WriteReturn() error {
	lines := []string{
		"// return",
		// frame = lcl
		"@LCL", // save current LCL to temp storage so it's not lost
		"D=M",
		"@R13",
		"M=D",

		// retAddr = *(frame - 5)
		"@5", // return address is LCL - 5
		"D=A",
		"@R13",
		"A=M-D", // Get LCL - 5 and save to the D register
		"D=M",
		"@R14",
		"M=D", // save return address to R14

		// *ARG = pop()
		"@SP", // pop off stack and save to D register, this is the return value
		"M=M-1",
		"A=M",
		"D=M",
		"@ARG",
		"A=M",
		"M=D", // Save the return value from the D register into the ARG address location

		// SP = ARG+1
		"@ARG",
		"D=M+1",
		"@SP",
		"M=D",

		// THAT = *(frame - 1)
		"@R13", // frame
		"A=M-1",
		"D=M",
		"@THAT",
		"M=D",

		// THIS = *(frame-2)
		"@2",
		"D=A",
		"@R13",
		"A=M-D",
		"D=M",
		"@THIS",
		"M=D",

		// ARG = *(frame - 3)
		"@3",
		"D=A",
		"@R13",
		"A=M-D",
		"D=M",
		"@ARG",
		"M=D",

		// LCL = *(frame - 4)
		"@4",
		"D=A",
		"@R13",
		"A=M-D",
		"D=M",
		"@LCL",
		"M=D",

		// goto retAddr
		"@R14",
		"A=M",
		"0;JMP",
	}

	return cw.writeLines(lines)
}
