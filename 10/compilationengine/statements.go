package compilationengine

import (
	"github.com/ctjnkns/nand2tetris/10/jackanalyzer/tokenizer"
)

func (ce *CompilationEngine) CompileStatements() error {
	if err := ce.writeLine("<statements>"); err != nil {
		return err
	}

	ce.indent++

	for ce.tokenizer.HasMoreTokens() && ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		switch kw {
		case tokenizer.LET:
			if err := ce.CompileLet(); err != nil {
				return err
			}
		case tokenizer.IF:
			if err := ce.CompileIf(); err != nil {
				return err
			}
		case tokenizer.WHILE:
			if err := ce.CompileWhile(); err != nil {
				return err
			}
		case tokenizer.DO:
			if err := ce.CompileDo(); err != nil {
				return err
			}
		case tokenizer.RETURN:
			if err := ce.CompileReturn(); err != nil {
				return err
			}
		}
	}

	ce.indent--

	return ce.writeLine("</statements>")
}

func (ce *CompilationEngine) CompileLet() error {
	if err := ce.writeLine("<letStatement>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeKeyword("let"); err != nil {
		return err
	}

	// varName
	if err := ce.writeIdentifier(); err != nil {
		return err
	}

	if err := ce.writeSymbol("="); err != nil {
		return err
	}

	if err := ce.CompileExpression(); err != nil {
		return err
	}

	if err := ce.writeSymbol(";"); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</letStatement>")
}

func (ce *CompilationEngine) CompileIf() error {
	if err := ce.writeLine("<ifStatement>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeKeyword("if"); err != nil {
		return err
	}

	if err := ce.writeSymbol("("); err != nil {
		return err
	}

	if err := ce.CompileExpression(); err != nil {
		return err
	}

	if err := ce.writeSymbol(")"); err != nil {
		return err
	}

	if err := ce.writeSymbol("{"); err != nil {
		return err
	}

	if err := ce.CompileStatements(); err != nil {
		return err
	}

	if err := ce.writeSymbol("}"); err != nil {
		return err
	}

	if ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw == tokenizer.ELSE {
			if err := ce.writeKeyword("else"); err != nil {
				return err
			}

			if err := ce.writeSymbol("{"); err != nil {
				return err
			}

			if err := ce.CompileStatements(); err != nil {
				return err
			}

			if err := ce.writeSymbol("}"); err != nil {
				return err
			}
		}
	}

	ce.indent--

	return ce.writeLine("</ifStatement>")
}

func (ce *CompilationEngine) CompileWhile() error {
	if err := ce.writeLine("<whileStatement>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeKeyword("while"); err != nil {
		return err
	}

	if err := ce.writeSymbol("("); err != nil {
		return err
	}

	if err := ce.CompileExpression(); err != nil {
		return err
	}

	if err := ce.writeSymbol(")"); err != nil {
		return err
	}

	if err := ce.writeSymbol("{"); err != nil {
		return err
	}

	if err := ce.CompileStatements(); err != nil {
		return err
	}

	if err := ce.writeSymbol("}"); err != nil {
		return err
	}

	ce.indent--

	if err := ce.writeLine("</whileStatement>"); err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) CompileDo() error {
	if err := ce.writeLine("<doStatement>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeKeyword("do"); err != nil {
		return err
	}

	if err := ce.compileSubroutineCall(); err != nil {
		return err
	}

	if err := ce.writeSymbol(";"); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</doStatement>")
}

func (ce *CompilationEngine) CompileReturn() error {
	if err := ce.writeLine("<returnStatement>"); err != nil {
		return err
	}

	ce.indent++

	if err := ce.writeKeyword("return"); err != nil {
		return err
	}

	if ce.tokenizer.Token() != ";" {
		if err := ce.CompileExpression(); err != nil {
			return err
		}
	}

	if err := ce.writeSymbol(";"); err != nil {
		return err
	}

	ce.indent--

	return ce.writeLine("</returnStatement>")
}
