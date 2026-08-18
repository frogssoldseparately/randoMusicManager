package parser

import (
	"fmt"
	"os"
	"strings"
	"github.com/dlclark/regexp2/v2"
)

type Parser struct {
	lineNumber int
	columnNumber int
}

const regexOptions = regexp2.Singleline

func MakeParser() *Parser {
	return &Parser{0, 0}
}

func (parser *Parser) incrementLineNumber() {
	parser.lineNumber++
	parser.columnNumber = 0
}

func (parser *Parser) ParseFile(fileName string) (*[]*Statement, error) {
	buf, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("Could not read %s : %s", fileName, err)
	}
	raw := strings.ReplaceAll(string(buf), "\r", "")
	var tokens []*Token
	tokenRunes := []rune{}
	for len(raw) > 0 {
		dLength, token := parser.readToken(&raw)
		if dLength == -1 {
			return nil, fmt.Errorf("Could not parse token at %d:%d", parser.lineNumber, parser.columnNumber)
		}
		if token != nil {
			tokens = append(tokens, token)
			tokenRunes = append(tokenRunes, rune(token.id))
		}
		raw = raw[dLength:]
		parser.columnNumber += dLength	}
	var expressions []*Expression
	expressionRunes := []rune{}
	for len(tokenRunes) > 0 {
		dLength, expression := parser.readExpression(&tokenRunes, &tokens)
		if dLength == -1 {
			return nil, fmt.Errorf("Could not parse expression at %d:%d", tokens[0].line, tokens[0].column)
		}
		if expression != nil {
			expressions = append(expressions, expression)
			expressionRunes = append(expressionRunes, rune(expression.id))
		}
		tokenRunes = tokenRunes[dLength:]
		tokens = tokens[dLength:]
	}
	var statements []*Statement
	for len(expressionRunes) > 0 {
		dLength, statement := parser.readStatement(&expressionRunes, &expressions)
		if dLength == -1 {
			return nil, fmt.Errorf("Could not parse statement at %d:%d", expressions[0].line, expressions[0].column)
		}
		if statement != nil {
			statements = append(statements, statement)
		}
		expressionRunes = expressionRunes[dLength:]
		expressions = expressions[dLength:]
	}
	return &statements, nil
}
