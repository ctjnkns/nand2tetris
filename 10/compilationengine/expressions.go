package compilationengine

import "github.com/ctjnkns/nand2tetris/10/jackanalyzer/tokenizer"

func (ce *CompilationEngine) CompileExpression() error {
	if err := ce.writeLine("<expression>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.CompileTerm(); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</expression>")
}

func (ce *CompilationEngine) CompileTerm() error {
	if err := ce.writeLine("<term>"); err != nil {
		return err
	}

	ce.indent++

	if ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		if err := ce.writeKeyword(""); err != nil {
			return err
		}
	} else {
		if err := ce.writeIdentifier(); err != nil {
			return err
		}
	}

	ce.indent--

	return ce.writeLine("</term>")
}

func (ce *CompilationEngine) compileSubroutineCall() error {
	if err := ce.writeIdentifier(); err != nil {
		return err
	}

	// handle . in subroutine call
	if ce.tokenizer.Token() == "." {
		if err := ce.writeSymbol("."); err != nil {
			return err
		}
		if err := ce.writeIdentifier(); err != nil {
			return err
		}
	}

	if err := ce.writeSymbol("("); err != nil {
		return err
	}

	if err := ce.CompileExpressionList(); err != nil {
		return err
	}

	if err := ce.writeSymbol(")"); err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) CompileExpressionList() error {
	if err := ce.writeLine("<expressionList>"); err != nil {
		return err
	}

	ce.indent++

	if ce.tokenizer.Token() != ")" {
		if err := ce.CompileExpression(); err != nil {
			return err
		}
		for ce.tokenizer.Token() == "," {
			if err := ce.writeSymbol(","); err != nil {
				return err
			}
			if err := ce.CompileExpression(); err != nil {
				return err
			}
		}
	}

	ce.indent--

	return ce.writeLine("</expressionList>")
}
