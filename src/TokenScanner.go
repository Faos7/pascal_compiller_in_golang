package main

import (
	"os"
	"bufio"
	"strings"
	"unicode/utf16"
)

type MyType byte

const (
	LETTER = iota
	DIGIT
	SPACE
	OPERATOR
	QUOTE
)

type TokenScanner struct {
	tokenName 	string
	lineRov 	int
	lineCol 	int
	readingString 	bool
	readingNumber 	bool
	isFloat 	bool
	sciNotation	bool
	readingColon	bool
	readingBool	bool
	readingDot	bool

	tokens 		[]Token

	KEYWORDS_TOKEN 	map[string] string

	OPERATORS_TOKEN map[string] string

	CHAR_TYPE map[string] MyType
}

func (ts* TokenScanner) clearStatuses() {
	ts.readingString = false
	ts.readingNumber = false
	ts.isFloat = false
	ts.sciNotation = false
	ts.readingColon = false
	ts.readingBool = false
}

func NewTokenScanner() *TokenScanner{
	return &TokenScanner{
		KEYWORDS_TOKEN : make(map[string] string),
		OPERATORS_TOKEN: make(map[string] string),
		CHAR_TYPE      : make(map[string] MyType),
	}
}

func (ts* TokenScanner) fillChars()  {

	for i:= 65; i<91; i++  {
		// add letters
		currChar := string(i)
		ts.CHAR_TYPE[currChar] = LETTER
		strings.ToLower(currChar)
		ts.CHAR_TYPE[currChar] = LETTER
	}

	for i:=48; i<58; i++ {
		//add digits
		currChar := string(i)
		ts.CHAR_TYPE[currChar] = DIGIT
	}

	for i:=1; i <33; i++ {
		//add spaces
		currChar := string(i)
		ts.CHAR_TYPE[currChar] = SPACE
	}

	for key := range ts.OPERATORS_TOKEN {
		ts.CHAR_TYPE[key] = OPERATOR

	}

	//add single quote
	ts.CHAR_TYPE[string(39)] = QUOTE
}

func (ts* TokenScanner) fillOps()  {
	ts.OPERATORS_TOKEN["("] = "TK_OPEN_PARENTHESIS"
	ts.OPERATORS_TOKEN[")"] = "TK_CLOSE_PARENTHESIS"
	ts.OPERATORS_TOKEN["["] = "TK_OPEN_SQUARE_BRACKET"
	ts.OPERATORS_TOKEN["]"] = "TK_CLOSE_SQUARE_BRACKET"

	ts.OPERATORS_TOKEN["."] = "TK_DOT"
	ts.OPERATORS_TOKEN[".."] = "TK_RANGE"
	ts.OPERATORS_TOKEN[":"] = "TK_COLON"
	ts.OPERATORS_TOKEN[";"] = "TK_SEMI_COLON"

	ts.OPERATORS_TOKEN["+"] = "TK_PLUS"
	ts.OPERATORS_TOKEN["-"] = "TK_MINUS"
	ts.OPERATORS_TOKEN["*"] = "TK_MULTIPLY"
	ts.OPERATORS_TOKEN["/"] = "TK_DIVIDE"

	ts.OPERATORS_TOKEN["<"] = "TK_LESS_THAN"
	ts.OPERATORS_TOKEN["<="] = "TK_LESS_THAN_EQUAL"
	ts.OPERATORS_TOKEN[">"] = "TK_GREATER_THAN"
	ts.OPERATORS_TOKEN[">="] = "TK_GREATER_THAN_EQUAL"

	ts.OPERATORS_TOKEN[":="] = "TK_ASSIGNMENT"
	ts.OPERATORS_TOKEN[","] = "TK_COMMA"
	ts.OPERATORS_TOKEN["="] = "TK_EQUAL"
	ts.OPERATORS_TOKEN["<>"] = "TK_NOT_EQUAL"

}

func (ts* TokenScanner) loadFromFIle ()  {
	file, err := os.Open("keywords.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var str string
	for scanner.Scan() {
		str = scanner.Text()
		ts.KEYWORDS_TOKEN[str] = strings.ToUpper(str)
	}
}

func (ts* TokenScanner) scan (code string) TokenList  {

	for i :=0;i<len(code) ;i++  {
		el := code[i];
		ts.checkCharacter(el)
	}

	ts.tokenName = "EOF"
	ts.generateToken("TK_EOF")

	res := NewTokenList(ts.tokens)
	return res

}

func (ts* TokenScanner) checkCharacter(element string){
	switch ts.CHAR_TYPE[element]{
	case LETTER:
		if !ts.readingNumber {
			ts.tokenName += element
		}

		if (element == 'E' && ts.readingNumber){
			ts.tokenName += element
			ts.sciNotation = true
		}
	case DIGIT:
		if (len(ts.tokenName) == 0) {
			ts.readingNumber = true
		}

		ts.tokenName += element

	case SPACE:
		if (ts.readingString) {
		//	append to string
			ts.tokenName += element
		} else if (ts.readingColon) {
			ts.generateToken(ts.OPERATORS_TOKEN[ts.tokenName])

			ts.readingColon = false
		} else if (ts.readingBool) {
			ts.generateToken(ts.OPERATORS_TOKEN[ts.tokenName])

			ts.readingBool = false
		} else if (! ts.readingNumber) {
		//	end of word
			ts.tokenName = ts.endOfWord()

			if element == utf16.Encode(10){
				// Check for newline on Unix OS
				ts.lineRov++
				ts.lineCol = 0
			} else if element == utf16.Encode(9){
				ts.lineCol+=4
			} else if element == utf16.Encode(32){
				ts.lineCol++
			}
		} else {
			ts.handleNumber()
		}
	case OPERATOR:
		if ts.readingDot && element == "." {
			if ts.tokenName == "." {
				ts.tokenName = ""
				ts.generateToken("TK_RANGE")
			} else {
				inputFmt := ts.tokenName[0:len(ts.tokenName)-2]
				ts.generateToken(inputFmt)
				ts.generateToken("TK_DOT")
				ts.tokenName = ""
			}
			ts.readingDot = false
		} else if ts.readingString {
		//	append to string
			ts.tokenName += element
		} else if ts.readingNumber {
			if ts.isFloat && element == "." {
				ts.isFloat = false
				inputFmt := ts.tokenName[0:len(ts.tokenName)-1]
				ts.tokenName = inputFmt
				ts.handleNumber()
				ts.generateToken("TK_RANGE")
				ts.tokenName = ""
			} else if ts.sciNotation && (element == "+" || element == "-"){
				ts.tokenName += element
			} else if element == "."{
			//	found decimal in float
				ts.isFloat=true
				ts.tokenName += element
			} else {
				ts.handleNumber()

				ts.generateToken(ts.OPERATORS_TOKEN[element])
			}
		} else if (ts.readingColon && element == "="){
			ts.tokenName += element
			ts.generateToken(ts.OPERATORS_TOKEN[ts.tokenName])
			ts.readingColon = false
		} else if (ts.readingBool){
			if ts.tokenName == "<" && ((element == "=") || element == ">") {
				ts.tokenName += element
				ts.generateToken(ts.OPERATORS_TOKEN[ts.tokenName])
			} else if ts.tokenName == ">" && element == "=" {
				ts.tokenName += element

				ts.generateToken(ts.OPERATORS_TOKEN[ts.tokenName])
			}

			ts.readingBool = false
		} else {
			if element == ";" {
			//	before end of line
				ts.tokenName = ts.endOfWord()
				ts.tokenName = ";"

				ts.generateToken(ts.OPERATORS_TOKEN[element])
			} else if element == ":"{
				ts.tokenName = ts.endOfWord()
				ts.readingColon = true
				ts.tokenName += element
			} else if element == "<" || element == ">"{
				ts.tokenName = ts.endOfWord()
				ts.readingColon = true
				ts.tokenName += element
			} else if element == "."{
				ts.tokenName += element

				if ts.tokenName == "end." {
					ts.generateToken("TK_END");
					ts.generateToken("TK_DOT");
				} else {
					ts.readingDot = true
				}
			} else if name, ok := ts.OPERATORS_TOKEN[element]; ok {
				ts.tokenName = ts.endOfWord()

				ts.tokenName = element

				ts.generateToken(name)
			}
		}
	case QUOTE:
	//	found begin/end quote
		ts.readingString = !ts.readingString
		ts.tokenName += element

		if !ts.readingString {
		//	remove trailing quotes
			inputFmt := ts.tokenName[1:len(ts.tokenName)-1]
			ts.tokenName = inputFmt

		//	found end quote
			if len(ts.tokenName) == 1 {
				ts.generateToken("TK_CHARLIT")
			} else if len(ts.tokenName) > 1 {
				ts.generateToken("TK_STRLIT");
			}
		}
	default:
		panic("Unhandled element scanned")
	}

}

func (ts* TokenScanner) handleNumber()  {
	ts.readingNumber = false
	if ts.isFloat {
		ts.generateToken("TK_FLOATLIT")
		ts.isFloat = false
	} else {
		ts.generateToken("TK_INTLIT")
	}
}

func (ts* TokenScanner) generateToken(tokenType string){
	t := NewToken(tokenType, ts.tokenName, ts.lineCol, ts.lineRov)

	ts.tokens = append(ts.tokens, t)

	ts.lineCol += len(ts.tokenName)

	ts.tokenName = ""
}

func (ts* TokenScanner) endOfWord() string  {
	if name, ok := ts.KEYWORDS_TOKEN[ts.tokenName]; ok {

		ts.generateToken(name)
	} else {
		if len(ts.tokenName) > 0 {
			if ts.tokenName == "true" || ts.tokenName == "false" {
				ts.generateToken("TK_BOOLLIT")
			} else {
				ts.generateToken("TK_IDENTIFIER")
			}
		}
	}

	ts.clearStatuses()
	return ts.tokenName
}
