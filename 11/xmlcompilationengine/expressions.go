package xmlcompilationengine

import "github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"

func (ce *CompilationEngine) CompileExpression() error {
	if err := ce.writeLine("<expression>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.CompileTerm(); err != nil {
		return err
	}

	for isOp(ce.tokenizer.Token()) {
		if err := ce.writeSymbol(ce.tokenizer.Token()); err != nil {
			return err
		}
		if err := ce.CompileTerm(); err != nil {
			return err
		}
	}

	ce.indent--

	return ce.writeLine("</expression>")
}

func isOp(s string) bool {
	switch s {
	case "+", "-", "*", "/", "&", "|", "<", ">", "=":
		return true
	}
	return false
}

func (ce *CompilationEngine) CompileTerm() error {
	if err := ce.writeLine("<term>"); err != nil {
		return err
	}

	ce.indent++

	switch ce.tokenizer.TokenType() {
	case tokenizer.INT_CONST:
		if err := ce.writeExpectedAndAdvance(tokenizer.INT_CONST, ""); err != nil {
			return err
		}
	case tokenizer.STRING_CONST:
		if err := ce.writeExpectedAndAdvance(tokenizer.STRING_CONST, ""); err != nil {
			return err
		}
	case tokenizer.KEYWORD:
		if err := ce.writeKeyword(""); err != nil {
			return err
		}
	case tokenizer.IDENTIFIER:
		name := ce.tokenizer.Token()

		switch ce.tokenizer.Peek() {
		case "[":
			if err := ce.writeVariableUse(name); err != nil {
				return err
			}
			if err := ce.writeSymbol("["); err != nil {
				return err
			}
			if err := ce.CompileExpression(); err != nil {
				return err
			}
			if err := ce.writeSymbol("]"); err != nil {
				return err
			}

		case "(":
			if err := ce.writeIdentifierInfo(name, "subroutine", "used", nil); err != nil {
				return err
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

		case ".":
			if _, ok := ce.lookup(name); ok {
				if err := ce.writeVariableUse(name); err != nil {
					return err
				}
			} else {
				if err := ce.writeIdentifierInfo(name, "class", "used", nil); err != nil {
					return err
				}
			}

			if err := ce.writeSymbol("."); err != nil {
				return err
			}

			subroutineName := ce.tokenizer.Token()
			if err := ce.writeIdentifierInfo(subroutineName, "subroutine", "used", nil); err != nil {
				return err
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

		default:
			if err := ce.writeVariableUse(name); err != nil {
				return err
			}
		}
	case tokenizer.SYMBOL:
		if ce.tokenizer.Token() == "(" {
			if err := ce.writeSymbol("("); err != nil {
				return err
			}
			if err := ce.CompileExpression(); err != nil {
				return err
			}
			if err := ce.writeSymbol(")"); err != nil {
				return err
			}
		} else {
			// unaryOp: - or ~
			if err := ce.writeSymbol(ce.tokenizer.Token()); err != nil {
				return err
			}
			if err := ce.CompileTerm(); err != nil {
				return err
			}
		}
	}

	ce.indent--

	return ce.writeLine("</term>")
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
