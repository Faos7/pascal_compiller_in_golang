package main

type TokenList struct {
	list []Token
	currNumber int
}

func NewTokenList( tokens []Token) *TokenList {
	return &TokenList{
		list : tokens,
		currNumber:-1,
	}
}

func (tl* TokenList) hasNext() bool  {
	return tl.currNumber < len(tl.list)
}

func (tl * TokenList) setIteratorToNull()  {
	tl.currNumber = -1
}

func (tl * TokenList) getNext() Token {
	if tl.currNumber == len(tl.list) {
		tl.currNumber = -1
	}
	tl.currNumber++
	return tl.list[tl.currNumber]
}
