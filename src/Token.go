package main

type Token struct {
	 tokenType string
	 tokenValue string

	 lineCol int
	 lineRow int
}

func NewToken(tt string, tv string, lc int, lr int)  *Token{
	return &Token{
		tokenType: 	tt,
		tokenValue: 	tv,
		lineCol:	lc,
		lineRow: 	lr,
	}
}

func (t* Token)getTokenVal() string {
	return t.tokenValue
}

func (t* Token)getTokenType() string  {
	return t.tokenType

}
func (t* Token)setTokenType(tt string)  {
	t.tokenType = tt
}

func (t* Token)toString() string  {
	return t.tokenValue
}
