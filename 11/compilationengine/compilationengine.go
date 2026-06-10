package compilationengine

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/symboltable"
	"github.com/ctjnkns/nand2tetris/11/jackanalyzer/tokenizer"
)

type Compiler interface {
	CompileClass() error
	CompileClassVarDec() error
	CompileSubroutine() error
	CompileParameterList() error
	CompileSubroutineBody() error
	CompileVarDec() error
	CompileStatements() error
	CompileLet() error
	CompileIf() error
	CompileWhile() error
	CompileDo() error
	CompileReturn() error
	CompileExpression() error
	CompileTerm() error
	CompileExpressionList() (int, error)
}

type CompilationEngine struct {
	tokenizer       *tokenizer.Tokenizer
	writer          *bufio.Writer
	className       string
	subroutineName  string
	classTable      *symboltable.SymbolTable
	subroutineTable *symboltable.SymbolTable
	subroutineKind  int
	labelIndex      int
	indent          int
}

func NewCompilationEngine(tokenizer *tokenizer.Tokenizer, writer *bufio.Writer) *CompilationEngine {
	ce := &CompilationEngine{
		tokenizer:       tokenizer,
		writer:          writer,
		classTable:      symboltable.NewSymbolTable(),
		subroutineTable: symboltable.NewSymbolTable(),
		indent:          0,
	}

	return ce
}

func (ce *CompilationEngine) writeVarDec() error {
	kw, err := ce.tokenizer.KeyWord()
	if err != nil {
		return err
	}

	var kind symboltable.Kind

	switch kw {
	case tokenizer.STATIC:
		kind = symboltable.STATIC
	case tokenizer.FIELD:
		kind = symboltable.FIELD
	case tokenizer.VAR:
		kind = symboltable.VAR
	default:
		return fmt.Errorf("expected static, field, or var; got %s", ce.tokenizer.Token())
	}

	// static|field|var
	if err := ce.consumeKeyword(""); err != nil {
		return err
	}

	typeName := ce.tokenizer.Token()

	// int|char|boolean|className
	if err := ce.consumeType(); err != nil {
		return err
	}

	name := ce.tokenizer.Token()

	var table *symboltable.SymbolTable
	if kind == symboltable.VAR {
		table = ce.subroutineTable
	} else {
		table = ce.classTable
	}

	if err := table.Define(name, typeName, kind); err != nil {
		return err
	}

	// varName
	if err := ce.consumeIdentifier(); err != nil {
		return err
	}

	// (, varName)*
	for ce.tokenizer.TokenType() == tokenizer.SYMBOL && ce.tokenizer.Token() == "," {
		if err := ce.consumeSymbol(","); err != nil {
			return err
		}

		name := ce.tokenizer.Token()

		if err := table.Define(name, typeName, kind); err != nil {
			return err
		}

		if err := ce.consumeIdentifier(); err != nil {
			return err
		}
	}

	return ce.consumeSymbol(";")
}

func (ce *CompilationEngine) writeLine(line string) error {
	indent := strings.Repeat("  ", ce.indent)
	_, err := ce.writer.WriteString(indent + line)
	if err != nil {
		return err
	}
	_, err = ce.writer.WriteRune('\n')
	if err != nil {
		return err
	}

	return nil
}

func (ce *CompilationEngine) checkAndAdvance() error {
	if ce.tokenizer.HasMoreTokens() {
		ce.tokenizer.Advance()
	}

	return nil
}

func (ce *CompilationEngine) writeVariableUse(name string) error {
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

	return ce.writeLine(fmt.Sprintf("push %s %d", segment, index))
}

func (ce *CompilationEngine) lookup(name string) (*symboltable.SymbolTable, bool) {
	if ce.subroutineTable.KindOf(name) != symboltable.NONE {
		return ce.subroutineTable, true
	}
	if ce.classTable.KindOf(name) != symboltable.NONE {
		return ce.classTable, true
	}
	return nil, false
}

func (ce *CompilationEngine) consumeIdentifier() error {
	if ce.tokenizer.TokenType() != tokenizer.IDENTIFIER {
		return fmt.Errorf("expected identifier; got %s", ce.tokenizer.Token())
	}
	return ce.checkAndAdvance()
}

func (ce *CompilationEngine) consumeSymbol(s string) error {
	if ce.tokenizer.TokenType() != tokenizer.SYMBOL || ce.tokenizer.Token() != s {
		return fmt.Errorf("expected symbol %s; got %s", s, ce.tokenizer.Token())
	}
	return ce.checkAndAdvance()
}

func (ce *CompilationEngine) consumeKeyword(s string) error {
	if ce.tokenizer.TokenType() != tokenizer.KEYWORD || (s != "" && ce.tokenizer.Token() != s) {
		return fmt.Errorf("expected keyword %s; got %s", s, ce.tokenizer.Token())
	}
	return ce.checkAndAdvance()
}

func (ce *CompilationEngine) consumeReturnType() error {
	if ce.tokenizer.TokenType() == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw == tokenizer.VOID {
			return ce.consumeKeyword("void")
		}
	}

	return ce.consumeType()
}

func (ce *CompilationEngine) consumeType() error {
	tt := ce.tokenizer.TokenType()
	if tt != tokenizer.KEYWORD && tt != tokenizer.IDENTIFIER {
		return fmt.Errorf("expected type; got: %s", ce.tokenizer.Token())
	}

	if tt == tokenizer.KEYWORD {
		kw, err := ce.tokenizer.KeyWord()
		if err != nil {
			return err
		}

		if kw != tokenizer.INT && kw != tokenizer.CHAR && kw != tokenizer.BOOLEAN {
			return fmt.Errorf("expected int|char|boolean; got %s", ce.tokenizer.Token())
		}
	}

	return ce.checkAndAdvance()
}

func (ce *CompilationEngine) writeOp(op string) error {
	switch op {
	case "+":
		return ce.writeLine("add")
	case "-":
		return ce.writeLine("sub")
	case "*":
		return ce.writeLine("call Math.multiply 2")
	case "/":
		return ce.writeLine("call Math.divide 2")
	case "&":
		return ce.writeLine("and")
	case "|":
		return ce.writeLine("or")
	case "<":
		return ce.writeLine("lt")
	case ">":
		return ce.writeLine("gt")
	case "=":
		return ce.writeLine("eq")
	default:
		return fmt.Errorf("unsupported op: %s", op)
	}
}

func kindSegment(kind symboltable.Kind) (string, error) {
	switch kind {
	case symboltable.STATIC:
		return "static", nil
	case symboltable.FIELD:
		return "this", nil
	case symboltable.ARG:
		return "argument", nil
	case symboltable.VAR:
		return "local", nil
	default:
		return "", fmt.Errorf("unsupported variable kind: %d", kind)
	}
}
