package codewriter

import "fmt"

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

func (cw *CodeWriter) WriteGoto(label string) error {
	lines := []string{
		fmt.Sprintf("// goto %s", label),
		fmt.Sprintf("@%s", label),
		"0;JMP",
	}

	return cw.writeLines(lines)
}
