package compilationengine

import (
	"fmt"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"
)

func (ce *CompilationEngine) CompileStatements() error {
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
	return nil
}

func (ce *CompilationEngine) CompileLet() error {
	if err := ce.consumeKeyword("let"); err != nil {
		return err
	}

	// varName
	name := ce.tokenizer.Token()
	table, ok := ce.lookup(name)
	if !ok {
		return fmt.Errorf("identifier not found: %s", name)
	}

	kind := table.KindOf(name)
	index := table.IndexOf(name)

	segment, err := kindSegment(kind)
	if err != nil {
		return err
	}

	if err := ce.consumeIdentifier(); err != nil {
		return err
	}

	isArrayAssignment := false
	if ce.tokenizer.Token() == "[" {
		isArrayAssignment = true
		if err := ce.consumeSymbol("["); err != nil {
			return err
		}

		if err := ce.CompileExpression(); err != nil {
			return err
		}

		if err := ce.consumeSymbol("]"); err != nil {
			return err
		}

		if err := ce.writeLine(fmt.Sprintf("push %s %d", segment, index)); err != nil {
			return err
		}
		if err := ce.writeLine("add"); err != nil {
			return err
		}
	}

	if err := ce.consumeSymbol("="); err != nil {
		return err
	}

	if err := ce.CompileExpression(); err != nil {
		return err
	}

	if err := ce.consumeSymbol(";"); err != nil {
		return err
	}

	if !isArrayAssignment {
		return ce.writeLine(fmt.Sprintf("pop %s %d", segment, index))
	}

	if err := ce.writeLine("pop temp 0"); err != nil {
		return err
	}

	if err := ce.writeLine("pop pointer 1"); err != nil {
		return err
	}

	if err := ce.writeLine("push temp 0"); err != nil {
		return err
	}

	return ce.writeLine("pop that 0")
}

func (ce *CompilationEngine) CompileIf() error {
	endLabel := ce.nextLabel()
	falseLabel := ce.nextLabel()

	if err := ce.consumeKeyword("if"); err != nil {
		return err
	}

	if err := ce.consumeSymbol("("); err != nil {
		return err
	}

	if err := ce.CompileExpression(); err != nil {
		return err
	}

	if err := ce.consumeSymbol(")"); err != nil {
		return err
	}

	if err := ce.writeLine("not"); err != nil {
		return err
	}

	if err := ce.writeLine(fmt.Sprintf("if-goto %s", falseLabel)); err != nil {
		return err
	}

	if err := ce.consumeSymbol("{"); err != nil {
		return err
	}

	if err := ce.CompileStatements(); err != nil {
		return err
	}

	if err := ce.consumeSymbol("}"); err != nil {
		return err
	}

	if err := ce.writeLine(fmt.Sprintf("goto %s", endLabel)); err != nil {
		return err
	}

	ce.indent = 0
	if err := ce.writeLine(fmt.Sprintf("label %s", falseLabel)); err != nil {
		return err
	}
	ce.indent = 2

	if ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw == tokenizer.ELSE {
			if err := ce.consumeKeyword("else"); err != nil {
				return err
			}

			if err := ce.consumeSymbol("{"); err != nil {
				return err
			}

			if err := ce.CompileStatements(); err != nil {
				return err
			}

			if err := ce.consumeSymbol("}"); err != nil {
				return err
			}
		}
	}

	ce.indent = 0
	if err := ce.writeLine(fmt.Sprintf("label %s", endLabel)); err != nil {
		return err
	}
	ce.indent = 2

	return nil
}

func (ce *CompilationEngine) nextLabel() string {
	label := fmt.Sprintf("%s_%d", ce.className, ce.labelIndex)
	ce.labelIndex++
	return label
}

func (ce *CompilationEngine) CompileWhile() error {
	startLabel := ce.nextLabel()
	endLabel := ce.nextLabel()

	ce.indent = 0
	if err := ce.writeLine(fmt.Sprintf("label %s", startLabel)); err != nil {
		return err
	}
	ce.indent = 2

	if err := ce.consumeKeyword("while"); err != nil {
		return err
	}

	if err := ce.consumeSymbol("("); err != nil {
		return err
	}

	if err := ce.CompileExpression(); err != nil {
		return err
	}

	if err := ce.consumeSymbol(")"); err != nil {
		return err
	}

	if err := ce.writeLine("not"); err != nil {
		return err
	}

	if err := ce.writeLine(fmt.Sprintf("if-goto %s", endLabel)); err != nil {
		return err
	}

	if err := ce.consumeSymbol("{"); err != nil {
		return err
	}

	if err := ce.CompileStatements(); err != nil {
		return err
	}

	if err := ce.consumeSymbol("}"); err != nil {
		return err
	}

	if err := ce.writeLine(fmt.Sprintf("goto %s", startLabel)); err != nil {
		return err
	}

	ce.indent = 0
	if err := ce.writeLine(fmt.Sprintf("label %s", endLabel)); err != nil {
		return err
	}
	ce.indent = 2

	return nil
}

func (ce *CompilationEngine) CompileDo() error {
	if err := ce.consumeKeyword("do"); err != nil {
		return err
	}

	if err := ce.CompileExpression(); err != nil {
		return err
	}

	if err := ce.consumeSymbol(";"); err != nil {
		return err
	}

	return ce.writeLine("pop temp 0") // discard return value
}

func (ce *CompilationEngine) CompileReturn() error {
	if err := ce.consumeKeyword("return"); err != nil {
		return err
	}

	if ce.tokenizer.Token() != ";" {
		if err := ce.CompileExpression(); err != nil {
			return err
		}
	} else {
		if err := ce.writeLine("push constant 0"); err != nil {
			return err
		}
	}

	if err := ce.consumeSymbol(";"); err != nil {
		return err
	}

	return ce.writeLine("return")
}
