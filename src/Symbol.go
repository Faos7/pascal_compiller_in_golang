package main


type Symbol struct {
	name		string
	tokenType	string
	dataType	PARSER_TYPE
	address		int32
	returnAddress	int32

	low		int //can be int or char
	high		int //can be int or char

	indexType	PARSER_TYPE
	valueType	PARSER_TYPE

	next 		*Symbol
}

func NewSymbol(name string, tokenType string, dataType PARSER_TYPE, address int32) *Symbol {
	return &Symbol{
		name	 :	name,
		tokenType:	tokenType,
		dataType :	dataType,
		address	 :	address,
	}
}

func (s* Symbol) getName() string  {
	return s.name
}

func (s* Symbol) getDataType() string  {
	return s.dataType
}

func (s* Symbol) getAddress() int32  {
	return s.address
}

func (s* Symbol) setAddress(add int32) {
	s.address = add
}

func (s* Symbol) getTokenType() string  {
	return s.tokenType
}

func (s* Symbol) setTokenType(tt string)  {
	s.tokenType = tt
}

func (s* Symbol) getReturnAddress() int  {
	return s.returnAddress
}

func (s* Symbol) setReturnAddress(ra int32)  {
	s.returnAddress = ra
}

func (s* Symbol) getLow() int  {
	return s.low
}

func (s* Symbol) setLow(lw int)  {
	s.low = lw
}

func (s* Symbol) getHigh() int  {
	return s.high
}

func (s* Symbol) setHigh(hg int) {
	s.high = hg
}

func (s* Symbol) getIndexType() PARSER_TYPE  {
	return s.indexType
}

func (s* Symbol) setIndexType(it PARSER_TYPE)  {
	s.indexType = it
}

func (s* Symbol) getValueType() PARSER_TYPE  {
	return  s.valueType
}

func (s* Symbol) setValueType(vt PARSER_TYPE)   {
	s.valueType = vt
}


