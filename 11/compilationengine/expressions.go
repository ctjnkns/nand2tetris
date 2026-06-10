package compilationengine

import (
	"fmt"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"
)

func (ce *CompilationEngine) CompileExpression() error {
	if err := ce.CompileTerm(); err != nil {
		return err
	}

	for isOp(ce.tokenizer.Token()) {
		// capture the op first
		op := ce.tokenizer.Token()
		if err := ce.consumeSymbol(op); err != nil {
			return err
		}

		if err := ce.CompileTerm(); err != nil {
			return err
		}

		if err := ce.writeOp(op); err != nil {
			return err
		}
	}

	return nil
}

func isOp(s string) bool {
	switch s {
	case "+", "-", "*", "/", "&", "|", "<", ">", "=":
		return true
	}
	return false
}

func (ce *CompilationEngine) CompileTerm() error {
	switch ce.tokenizer.TokenType() {
	case tokenizer.INT_CONST:
		value := ce.tokenizer.Token()
		if err := ce.writeLine(fmt.Sprintf("push constant %s", value)); err != nil {
			return err
		}

		if err := ce.checkAndAdvance(); err != nil {
			return err
		}
	case tokenizer.STRING_CONST:
		value := ce.tokenizer.Token()

		if err := ce.writeLine(fmt.Sprintf("push constant %d", len(value))); err != nil {
			return err
		}
		if err := ce.writeLine("call String.new 1"); err != nil {
			return err
		}

		for _, ch := range value {
			if err := ce.writeLine(fmt.Sprintf("push constant %d", ch)); err != nil {
				return err
			}

			if err := ce.writeLine("call String.appendChar 2"); err != nil {
				return err
			}
		}

		if err := ce.checkAndAdvance(); err != nil {
			return err
		}
	case tokenizer.KEYWORD:
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		switch kw {
		case tokenizer.TRUE:
			if err := ce.writeLine("push constant 1"); err != nil {
				return err
			}
			if err := ce.writeLine("neg"); err != nil {
				return err
			}
		case tokenizer.FALSE, tokenizer.NULL:
			if err := ce.writeLine("push constant 0"); err != nil {
				return err
			}
		case tokenizer.THIS:
			if err := ce.writeLine("push pointer 0"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported keyword constant: %s", ce.tokenizer.Token())
		}

		if err := ce.checkAndAdvance(); err != nil {
			return err
		}
	case tokenizer.IDENTIFIER:
		name := ce.tokenizer.Token()

		switch ce.tokenizer.Peek() {
		case "[":
			if err := ce.consumeIdentifier(); err != nil {
				return err
			}

			if err := ce.consumeSymbol("["); err != nil {
				return err
			}
			if err := ce.CompileExpression(); err != nil {
				return err
			}
			if err := ce.consumeSymbol("]"); err != nil {
				return err
			}

			table, ok := ce.lookup(name)
			if !ok {
				return fmt.Errorf("identifier not found: %s", name)
			}
			segment, err := kindSegment(table.KindOf(name))
			if err != nil {
				return err
			}

			if err := ce.writeLine(fmt.Sprintf("push %s %d", segment, table.IndexOf(name))); err != nil {
				return err
			}
			if err := ce.writeLine("add"); err != nil {
				return err
			}

			if err := ce.writeLine("pop pointer 1"); err != nil {
				return err
			}
			return ce.writeLine("push that 0")

		case "(":
			if err := ce.writeLine("push pointer 0"); err != nil {
				return err
			}

			if err := ce.consumeIdentifier(); err != nil {
				return err
			}
			if err := ce.consumeSymbol("("); err != nil {
				return err
			}
			nArgs, err := ce.CompileExpressionList()
			if err != nil {
				return err
			}
			if err := ce.consumeSymbol(")"); err != nil {
				return err
			}
			if err := ce.writeLine(fmt.Sprintf("call %s.%s %d", ce.className, name, nArgs+1)); err != nil {
				return err
			}

		case ".":
			callName := name
			nExtraArgs := 0

			if table, ok := ce.lookup(name); ok {
				segment, err := kindSegment(table.KindOf(name))
				if err != nil {
					return err
				}
				if err := ce.writeLine(fmt.Sprintf("push %s %d", segment, table.IndexOf(name))); err != nil {
					return err
				}
				callName = table.TypeOf(name)
				nExtraArgs = 1
			}

			if err := ce.consumeIdentifier(); err != nil {
				return err
			}

			if err := ce.consumeSymbol("."); err != nil {
				return err
			}

			subroutineName := ce.tokenizer.Token()
			if err := ce.consumeIdentifier(); err != nil {
				return err
			}

			if err := ce.consumeSymbol("("); err != nil {
				return err
			}

			nArgs, err := ce.CompileExpressionList()
			if err != nil {
				return err
			}

			if err := ce.consumeSymbol(")"); err != nil {
				return err
			}

			if err := ce.writeLine(fmt.Sprintf("call %s.%s %d", callName, subroutineName, nArgs+nExtraArgs)); err != nil {
				return err
			}

		default:
			if err := ce.writeVariableUse(name); err != nil {
				return err
			}
		}
	case tokenizer.SYMBOL:
		if ce.tokenizer.Token() == "(" {
			if err := ce.consumeSymbol("("); err != nil {
				return err
			}
			if err := ce.CompileExpression(); err != nil {
				return err
			}
			if err := ce.consumeSymbol(")"); err != nil {
				return err
			}
		} else {
			// unaryOp: - or ~
			op := ce.tokenizer.Token()

			if err := ce.consumeSymbol(ce.tokenizer.Token()); err != nil {
				return err
			}
			if err := ce.CompileTerm(); err != nil {
				return err
			}

			switch op {
			case "-":
				return ce.writeLine("neg")
			case "~":
				return ce.writeLine("not")
			default:
				return fmt.Errorf("unsupported unary op: %s", op)
			}
		}
	}

	return nil
}

func (ce *CompilationEngine) CompileExpressionList() (int, error) {
	nArgs := 0

	if ce.tokenizer.Token() != ")" {
		if err := ce.CompileExpression(); err != nil {
			return 0, err
		}

		nArgs++

		for ce.tokenizer.Token() == "," {
			if err := ce.consumeSymbol(","); err != nil {
				return 0, err
			}
			if err := ce.CompileExpression(); err != nil {
				return 0, err
			}
			nArgs++
		}
	}

	return nArgs, nil
}
