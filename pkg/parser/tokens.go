package parser

import (
	"strings"

	"github.com/dlclark/regexp2/v2"
)

type Token struct {
	id int
	value string
	line int
	column int
}

const (
	commentId int = 4
	rightBraceId int = 7
	whitespaceId int = 10
	stringId int = 5
	leftBraceId int = 6
	wordId int = 8
	newlineId int = 9
	varGetId int = 0
	varSetId int = 1
	varRefId int = 2
	funcNameId int = 3
)

var tokenMatchers = map[int]*regexp2.Regexp{
	stringId: regexp2.MustCompile(`^((\"[^\"]*?\"))`, regexOptions),
	leftBraceId: regexp2.MustCompile(`^(({))`, regexOptions),
	wordId: regexp2.MustCompile(`^((\S+))`, regexOptions),
	newlineId: regexp2.MustCompile(`^((\r?\n))`, regexOptions),
	varGetId: regexp2.MustCompile(`^((\%[a-zA-Z]+))`, regexOptions),
	varSetId: regexp2.MustCompile(`^((\:[a-zA-Z]+))`, regexOptions),
	varRefId: regexp2.MustCompile(`^((\$[a-zA-Z]+))`, regexOptions),
	funcNameId: regexp2.MustCompile(`^((\#[a-zA-Z]+))`, regexOptions),
	commentId: regexp2.MustCompile(`^((\/\/.*?(?=\r?\n)))`, regexOptions),
	rightBraceId: regexp2.MustCompile(`^((}))`, regexOptions),
	whitespaceId: regexp2.MustCompile(`^((\s+))`, regexOptions),
}

func (parser *Parser) makeToken(id int, components []string) *Token {
	ignore := false
	empty := false
	value := ""
	trimChars := ""
	joinChars := ""
	leftTrim := 0
	rightTrim := 0
	bothTrim := 0

	switch id {
	case varGetId:
		leftTrim = 1
	case varSetId:
		leftTrim = 1
	case varRefId:
		leftTrim = 1
	case funcNameId:
		leftTrim = 1
	case leftBraceId:
		empty = true
	case rightBraceId:
		empty = true
	case newlineId:
		empty = true
		parser.incrementLineNumber()
	case whitespaceId:
		ignore = true
	case commentId:
		ignore = true
	case stringId:
		trimChars = `"`
	default:
		// do nothing
	}

	if ignore {
		return nil
	}
	if !empty && len(value) == 0 {
		value = strings.Trim(strings.Join(components, joinChars), trimChars)
		if bothTrim != 0 {
			leftTrim = bothTrim
			rightTrim = bothTrim
		}
		if leftTrim+rightTrim+bothTrim > 0 {
			value = value[leftTrim : len(value) - rightTrim]
		}
	}
	return &Token{id, value, parser.lineNumber, parser.columnNumber}
}

func (parser *Parser) readToken(raw *string) (int, *Token) {
	id := -1
	matchLength := -1
	components := []string{}
	for key := 0; key <= whitespaceId; key++ {
		regex := tokenMatchers[key]
		if match, _ := regex.FindStringMatch(*raw); match != nil {
			id = key
			matchLength = match.RuneLength
			for _, component := range match.Captures {
				components = append(components, component.String())
			}
			break
		}
	}
	if id == -1 {
		return matchLength, nil
	}
	return matchLength, parser.makeToken(id, components)
}
