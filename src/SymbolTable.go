package main

var HASH_TABLE_SIZE = 211

type Scope struct {
	symbolTable 	[]Symbol
	next		*Scope
}

func NewScope() *Scope  {
	return &Scope{
		symbolTable : make([]Symbol, HASH_TABLE_SIZE),
	}
}

type SymbolTable struct {
	headerScope Scope
}

func NewSymbolTable() *SymbolTable  {
	return &SymbolTable{
		headerScope:NewScope(),
	}
}

func (st* SymbolTable) insert(symbol *Symbol)  {
	hashVal := st.hash(symbol.getName())
	bucketCursor := st.headerScope.symbolTable[hashVal]
	if bucketCursor == nil {
	//empty bucket
		st.headerScope.symbolTable[hashVal] = symbol
	} else {
		// Existing Symbols in bucket
		for bucketCursor.next != nil {
			bucketCursor = bucketCursor.next
		}

		// Append symbol at the end of the bucket
		bucketCursor.next = symbol;
	}
}

func (st* SymbolTable) lookup(symbolName string) Symbol{
	hashVal := st.hash(symbolName)
	bucketCursor := st.headerScope.symbolTable[hashVal]
	scopeCursor := st.headerScope

	for scopeCursor != nil{
		for bucketCursor != nil {
			if bucketCursor.name == symbolName {
				return bucketCursor
			}
			bucketCursor = bucketCursor.next
		}
		scopeCursor = scopeCursor.next
	}
//	symbol does not exist
	return nil
}

func (st* SymbolTable) hash(symbolName string) int {
	h :=0
	for i := 0; i < len(symbolName); i++ {
		h = h + h +symbolName[i]
	}

	h = h % HASH_TABLE_SIZE

	return h
}

func (st* SymbolTable) openScope(){
	innerScope := NewScope()

	innerScope.next = st.headerScope

	st.headerScope = innerScope
}

func (st* SymbolTable) closeScope()  {
	st.headerScope = st.headerScope.next
}

func (st* SymbolTable) getHeaderScope() Scope{
	return st.headerScope
}