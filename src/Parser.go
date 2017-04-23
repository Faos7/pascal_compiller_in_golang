package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"strconv"
)

type PARSER_TYPE byte
const  (
	_ = iota
	I
	R
	B
	C
	S
	P
	L
	A
)

type OP_CODE int
const (
	//_ = iota
	PUSHI = iota
	PUSH
	POP
	PUSHF

	JMP
	JFALSE
	JTRUE

	CVR
	CVI

 	DUP
	XCHG
	REMOVE


	ADD
	SUB
	MULT
	DIV
	NEG

	OR
	AND

	FADD
	FSUB
	FMULT
	FDIV
	FNEG

	EQL
	NEQL
	GEQ
	LEQ
	GTR
	LSS

	FGTR
	FLSS

	HALT

	PRINT_INT
	PRINT_CHAR
	PRINT_BOOL
	PRINT_REAL
	PRINT_NEWLINE

	GET
	PUT
)

var OP_CODE_STR = [...] string{
	"PUSHI", "PUSH", "POP", "PUSHF",

	"JMP", "JFALSE", "JTRUE",

	"CVR","CVI",

	"DUP", "XCHG", "REMOVE",

	"ADD", "SUB", "MULT", "DIV", "NEG",

	"OR", "AND",

	"FADD", "FSUB", "FMULT", "FDIV", "FNEG",

	"EQL", "NEQL", "GEQ", "LEQ", "GTR", "LSS",

	"FGTR", "FLSS",

	"HALT",

	"PRINT_INT", "PRINT_CHAR", "PRINT_BOOL", "PRINT_REAL", "PRINT_NEWLINE",

	"GET", "PUT",
}

type Parser struct {
	dp 		int32
	AddressSize 	int32
	currentToken 	Token
	INSTRUCTION_SIZE int32
	byteArray	[]byte
	ip 		int32
	STRING_TYPE_MAP	map[string] PARSER_TYPE
	tokens 		TokenList
	symbolTable 	*SymbolTable


}

func NewParser(tokens TokenList) *Parser  {
	return &Parser {
		dp 		: 0,
		AddressSize	: 4,
		INSTRUCTION_SIZE:1000,
		STRING_TYPE_MAP :make(map[string] PARSER_TYPE),
		//byteArray:
		ip		: 0,
		tokens		:tokens,


	}
}

func (p* Parser)initSTRING_TYPE_MAP(){
	p.STRING_TYPE_MAP["integer"]  = PARSER_TYPE(I)
	p.STRING_TYPE_MAP["real"] = PARSER_TYPE(R)
	p.STRING_TYPE_MAP["boolean"] = PARSER_TYPE(B)
	p.STRING_TYPE_MAP["char"] = PARSER_TYPE(C)
	p.STRING_TYPE_MAP["string"] = PARSER_TYPE(S)
	p.STRING_TYPE_MAP["array"] = PARSER_TYPE(A)
	//p.byteArray[p.INSTRUCTION_SIZE]
	p.symbolTable = NewSymbolTable()
}


func (p *Parser) parse () []byte  {
	p.getToken()

	p.match("TK_PROGRAM");
	p.match("TK_IDENTIFIER");
	p.match("TK_SEMI_COLON");

	p.program();

	return p.byteArray;
}

/*
    <pascal program> ->
	    [<program stat>]
	    <declarations>
	    <begin-statement>.
    <program stat> -> E
     */
func (p *Parser)program()  {
	p.declarations()
	p.begin()
}

/*
   <declarations> ->
	   <var decl><declarations>
	   <label ______,,______>
	   <type ______,,______>
	   <const ______,,______>
	   <procedure ______,,______>
	   <function ______,,______>
       -> E */
func (p* Parser) declarations()  {
       for(true) {
	       switch (p.currentToken.getTokenType()) {
	       case "TK_VAR":
		       p.varDeclarations();
		       break;
	       case "TK_PROCEDURE":
		       p.procDeclaration();
		       break;
	       case "TK_LABEL":
		       p.labelDeclarations();
		       break;
	       case "TK_BEGIN":
		       return;
	       }
       }
}

// label <namelist>;
func (p *Parser) labelDeclarations() {
       for(true) {
	       if ("TK_LABEL" == (p.currentToken.getTokenType())) {
		       p.match("TK_LABEL");
	       } else {
		       // currentToken is not "TK_LABEL"
		       break;
	       }

	       // Store labels in a list
	       //ArrayList<Token> labelsArrayList = new ArrayList<>();


	       var labelsArrayList []Token
	       for ("TK_IDENTIFIER" == (p.currentToken.getTokenType())) {
		       p.currentToken.setTokenType("TK_A_LABEL");
		       labelsArrayList = append(labelsArrayList, p.currentToken)

		       p.match("TK_A_LABEL");

		       if ("TK_COMMA"==p.currentToken.getTokenType()){
			       p.match("TK_COMMA");
		       }
	       }

	       // insert all labels into SymbolTable
	       //for (Token label : labelsArrayList) {

	       for i:=0; i < len(labelsArrayList); i++{
		       label := labelsArrayList[i]
	       symbol := NewSymbol(label.getTokenVal(), "TK_A_LABEL", PARSER_TYPE(L), 0);

		       ps := p.symbolTable.lookup(label.getTokenVal())
		       if (ps == nil) {
		       		p.symbolTable.insert(symbol);
		       }
	       }

	       p.match("TK_SEMI_COLON");
       }
}


/*
  <procedure decl> -> procedure <name> [params];
      <declarations>
      <begin-statement>
	  <statement> -> <procedure call>
  */
func (p *Parser) procDeclaration() {
	// declaration
	if (p.currentToken.getTokenType() == "TK_PROCEDURE") {
		p.match("TK_PROCEDURE");
		p.currentToken.setTokenType("TK_A_PROC");

		procedureName := p.currentToken.getTokenVal();

		p.match("TK_A_PROC");
		p.match("TK_SEMI_COLON");

		// generate hole to jump past the body
		p.genOpCode(OP_CODE(JMP));
		var hole int32 = p.ip;
		p.genAddressInt(0);

		symbol := NewSymbol(procedureName, "TK_A_PROC", PARSER_TYPE(P), p.ip);

		// body
		p.match("TK_BEGIN");
		p.statements();
		p.match("TK_END");
		p.match("TK_SEMI_COLON");

		// hole to return the procedure
		p.genOpCode(OP_CODE(JMP));
		symbol.setReturnAddress(p.ip);
		p.genAddressInt(0);

		if (p.symbolTable.lookup(procedureName) == nil) {
			p.symbolTable.insert(symbol);
		}

		// fill in the hole to jump past the body
		var save int32 = p.ip;

		p.ip = hole;
		p.genAddressInt(save);
		p.ip = save;
	}
}

/*
    <var decl> ->
        var[<namelist>: <type>;]^+
     */
func (p *Parser) varDeclarations() {
	for(true) {
		if ("TK_VAR" == (p.currentToken.getTokenType())) {
			p.match("TK_VAR");
		} else {
			// currentToken is not "TK_VAR"
			break;
		}

		// Store variables in a list

		var variablesArrayList []Token
		for ("TK_IDENTIFIER" == (p.currentToken.getTokenType())) {
			p.currentToken.setTokenType("TK_A_VAR");
			variablesArrayList = append(variablesArrayList, p.currentToken);

			p.match("TK_A_VAR");

			if ("TK_COMMA" == (p.currentToken.getTokenType())) {
				p.match("TK_COMMA");
			}
		}

		p.match("TK_COLON");
		dataType := p.currentToken.getTokenType();
		p.match(dataType);

		// Add the correct datatype for each identifier and insert into symbol table

		for i := 0; i < len(variablesArrayList); i++ {

			var variable Token = variablesArrayList[i]
			dt := strings.ToLower(dataType)

			symbol := NewSymbol(variable.getTokenVal(), "TK_A_VAR",
				p.STRING_TYPE_MAP[dt[3 : len(dt)]], p.dp);

			p.dp += 4;


			if (p.symbolTable.lookup(variable.getTokenVal()) == nil) {
				p.symbolTable.insert(symbol);
			}
		}

		if (dataType == ("TK_ARRAY")){
			p.arrayDeclaration(variablesArrayList);
		}

		p.match("TK_SEMI_COLON");

	}
}


/*
    <var decl> -> var <namelist>: <type>
    	<type> -> integer | real | bool | char
				<array type>
	    <array type> -> array[<low>..<high>]
				of <type> (simple type, array not allowed)


	    <low>,<high> ->
            ordinal constants of the same type
     */
func (p *Parser) arrayDeclaration(variablesArrayList []Token)  {
	p.match("TK_OPEN_SQUARE_BRACKET");
	v1 := p.currentToken.getTokenVal();
	indexType1 := p.getLitType(p.currentToken.getTokenType());
	p.match(p.currentToken.getTokenType());

	p.match("TK_RANGE");

	v2 := p.currentToken.getTokenVal();
	indexType2 := p.getLitType(p.currentToken.getTokenType());
	p.match(p.currentToken.getTokenType());
	p.match("TK_CLOSE_SQUARE_BRACKET");
	p.match("TK_OF");

	valueType := p.currentToken.getTokenType();
	p.match(valueType);

	if (indexType1 != indexType2){
		var s1 string = OP_CODE_STR[indexType1]
		var s2 string = OP_CODE_STR[indexType2]
		panic("Array index LHS type (" + s1 + ") is not equal to RHS type: (" + s2 + ")" );
	} else {

		//assert indexType1 != null;
		switch (indexType1) {
		case I:
			val1, _:= strconv.Atoi(v1)
			val2, _:= strconv.Atoi(v2)
			//var i1 int32 = val1;
			//var i2 int32 = val2;
			if (val1 > val2){
				panic("Array range is invalid: " + strconv.Itoa(val1) + ".." + strconv.Itoa(val2));
			}

			firstIntArray := p.symbolTable.lookup(variablesArrayList[0].getTokenVal());
			if (firstIntArray != nil) {
				p.dp = firstIntArray.getAddress();
			}

			for i := 0; i < len(variablesArrayList); i++ {


				variable := variablesArrayList[i];
				symbol := p.symbolTable.lookup(variable.getTokenVal());
				if (symbol != nil){

					var elementSize int = 4;
					var size int= elementSize*(val2 - val1 + 1);

					symbol.setAddress(p.dp);
					symbol.setLow(val1);
					symbol.setHigh(val2);
					symbol.setTokenType("TK_AN_ARRAY");
					symbol.setIndexType(PARSER_TYPE(I));
					vt := strings.ToLower(valueType);
					symbol.setValueType(p.STRING_TYPE_MAP[vt[3 : len(vt)]]);

					p.dp += int32(size);
				}
			}

		case C:
			a1 := []rune(v1)
			a2 := []rune(v2)
			var c1 rune = a1[0];
			var c2 rune = a2[0];
			if (c1 > c2){
				panic("Array range is invalid: " + strconv.Itoa(int(c1)) + ".." + strconv.Itoa(int(c2)));
			}

			firstCharArray := p.symbolTable.lookup(variablesArrayList[0].getTokenVal());
			if (firstCharArray != nil) {
				p.dp = firstCharArray.getAddress();
			}

			for i := 0; i < len(variablesArrayList); i++ {

			var variable Token = variablesArrayList[i]
				symbol := p.symbolTable.lookup(variable.getTokenVal());
				if (symbol != nil){
					size := c2 - c1 + 1;

					symbol.setAddress(p.dp);
					symbol.setLow(int(c1));
					symbol.setHigh(int(c2));
					symbol.setTokenType("TK_AN_ARRAY");
					symbol.setIndexType(PARSER_TYPE(C));

					vt := strings.ToLower(valueType);
					symbol.setValueType(p.STRING_TYPE_MAP[vt[3 : len(vt)]]);

					p.dp += size;
				}
			}
		case R:
			panic("Array index type: real is invalid");
		}

	}

}


/*
    <begin_statement> ->
        begin <stats> end
     */
func (p *Parser) begin()  {
	p.match("TK_BEGIN");
	p.statements();
	p.match("TK_END");
	p.match("TK_DOT");
	p.match("TK_EOF");
	p.genOpCode(OP_CODE(HALT));
}


/*
    <stats> ->
	    <while stat>; <stats>
	    <repeat ...
	    <goto ...
	    <case ...
	    <if ...
	    <for ...
	    <assignment> TK_A_VAR
	    <labelling> TK_A_LABEL
	    <procedure call> TK_A_PROC
	    <writeStat>
     */
func (p *Parser)statements() {
	for(! (p.currentToken.getTokenType() == "TK_END")) {
		switch (p.currentToken.getTokenType()) {
		case "TK_CASE":
			p.caseStat();
			break;
		case "TK_GOTO":
			p.goToStat();
			break;
		case "TK_WHILE":
			p.whileStat();
			break;
		case "TK_REPEAT":
			p.repeatStat();
			break;
		case "TK_IF":
			p.ifStat();
			break;
		case "TK_FOR":
			p.forStat();
			break;
		case "TK_WRITELN":
			p.writeStat();
			break;
		case "TK_IDENTIFIER":
			symbol := p.symbolTable.lookup(p.currentToken.getTokenVal());
			if (symbol != nil) {
				// assign token type to be var, proc, or label
				p.currentToken.setTokenType(symbol.getTokenType());
			}
			break;
		case "TK_A_VAR":
			p.assignmentStat();
			break;
		case "TK_A_PROC":
			p.procedureStat();
			break;
		case "TK_A_LABEL":
			p.labelStat();
			break;
		case "TK_AN_ARRAY":
			p.arrayAssignmentStat();
			break;
		case "TK_SEMI_COLON":
			p.match("TK_SEMI_COLON");
			break;
		default:
			return;
		}
	}
}

func (p * Parser) labelStat()  {
	symbol :=p.symbolTable.lookup(p.currentToken.getTokenVal());
	p.match("TK_A_LABEL");
	p.match("TK_COLON");
	if (symbol != nil) {
		hole := symbol.getAddress();
		save := p.ip;

		// fill in hole for goto jump
		p.ip = hole;
		p.genAddressInt(save);

		p.ip = save;

		p.statements();
	}
}

func (p * Parser) procedureStat()  {
	symbol := p.symbolTable.lookup(p.currentToken.getTokenVal());
	if (symbol != nil) {
		address := symbol.getAddress();
		p.match("TK_A_PROC");
		p.match("TK_SEMI_COLON");
		// call procedure
		p.genOpCode(OP_CODE(JMP));
		p.genAddressInt(address);

		restore := p.ip;

		// fill in return hole and restore ip
		p.ip = symbol.getReturnAddress();
		p.genAddressInt(restore);
		p.ip = restore;
	}
}

func (p *Parser)goToStat()  {
	p.match("TK_GOTO");
	symbol := p.symbolTable.lookup(p.currentToken.getTokenVal());
	p.currentToken.setTokenType("TK_A_LABEL");
	p.match("TK_A_LABEL");
	p.genOpCode(OP_CODE(JMP));
	hole := p.ip;
	p.genAddressInt(0);

	// hole for jump
	if (symbol != nil){
		symbol.setAddress(hole);
	}

	p.match("TK_SEMI_COLON");
}

// for <variable name> := <initial value> to <final value> do <stat>
func (p *Parser) forStat()  {
	p.match("TK_FOR");

	varName := p.currentToken.getTokenVal();
	p.currentToken.setTokenType("TK_A_VAR");
	p.assignmentStat();

	target := p.ip;


	symbol := p.symbolTable.lookup(varName);
	if (symbol != nil) {
		address := symbol.getAddress();
		p.match("TK_TO");

		// Generate op code for x <= <upper bound>
		p.genOpCode(OP_CODE(PUSH));
		p.genAddressInt(address);
		p.genOpCode(OP_CODE(PUSHI));
		p.genAddressInt(p.currentToken.getTokenVal());

		p.genOpCode(OP_CODE(LEQ));
		p.match("TK_INTLIT");

		p.match("TK_DO");

		p.genOpCode(OP_CODE(JFALSE));
		hole := p.ip;
		p.genAddressInt(0);

		p.match("TK_BEGIN");
		p.statements();
		p.match("TK_END");
		p.match("TK_SEMI_COLON");

		// Generate op code for x := x + 1;
		p.genOpCode(OP_CODE(PUSH));
		p.genAddressInt(address);
		p.genOpCode(OP_CODE(PUSHI));
		p.genAddressInt(1);
		p.genOpCode(OP_CODE(ADD));

		p.genOpCode(OP_CODE(POP));
		p.genAddressInt(address);


		p.genOpCode(OP_CODE(JMP));
		p.genAddressInt(target);

		save := p.ip;
		p.ip = hole;
		p.genAddressInt(save);
		p.ip = save;
	}
}


// repeat <stat> until <cond>
func (p * Parser)repeatStat()  {
	p.match("TK_REPEAT");
	target := p.ip;
	p.statements();
	p.match("TK_UNTIL");
	p.C();
	p.genOpCode(OP_CODE(JFALSE));
	p.genAddressInt(target);
}

// while <cond> do <stat>
func (p *Parser)whileStat()  {
	p.match("TK_WHILE");
	target := p.ip;
	p.C();
	p.match("TK_DO");

	p.genOpCode(OP_CODE(JFALSE));
	hole := p.ip;
	p.genAddressInt(0);

	p.match("TK_BEGIN");
	p.statements();
	p.match("TK_END");
	p.match("TK_SEMI_COLON");


	p.genOpCode(OP_CODE(JMP));
	p.genAddressInt(target);

	save := p.ip;
	p.ip = hole;
	p.genAddressInt(save);
	p.ip = save;
}

// if <cond> then <stat>
// if <cond> then <stat> else <stat>
func (p *Parser)ifStat()  {
	p.match("TK_IF");
	p.C();
	p.match("TK_THEN");
	p.genOpCode(OP_CODE(JFALSE));
	hole1 := p.ip;
	p.genAddressInt(0); // Holder value for the address
	p.statements();

	if(p.currentToken.getTokenType() == "TK_ELSE") {
		p.genOpCode(OP_CODE(JMP));
		hole2 := p.ip;
		p.genAddressInt(0);
		save := p.ip;
		p.ip = hole1;
		p.genAddressInt(save); // JFALSE to this else statement
		p.ip = save;
		hole1 = hole2;
		p.statements();
		p.match("TK_ELSE");
		p.statements();
	}

	save := p.ip;
	p.ip = hole1;
	p.genAddressInt(save); // JFALSE to outside the if statement in if-then or JMP past the else statement in if-else
	p.ip = save;
}

/*
    case E of
        [<tags>: <statement>]^+
            [else <statement>]

    end

    <tags> -> <single tag> 10:
          <range> 3..9:
          <list> 3,5,7:
          <list of ranges> 1..2,30..40:
    */
func (p *Parser) caseStat()  {
	p.match("TK_CASE");
	p.match("TK_OPEN_PARENTHESIS");
	eToken := p.currentToken;

	t1 := p.E();

	if (t1 == PARSER_TYPE(R)) {
		panic("Invalid type of real for case E");
	}

	p.match("TK_CLOSE_PARENTHESIS");
	p.match("TK_OF");

	var labelsArrayList []int ;

	for(p.currentToken.getTokenType() == "TK_INTLIT" ||
		p.currentToken.getTokenType() == "TK_CHARLIT" ||
		p.currentToken.getTokenType() == "TK_BOOLLIT") {

		t2 := p.E();
		p.emit("TK_EQUAL", t1, t2);
		p.match("TK_COLON");

		// hole for JFALSE to the next case label when the eql condition fails
		p.genOpCode(OP_CODE(JFALSE));
		hole := p.ip;
		p.genAddressInt(0);
		p.statements();

		p.genOpCode(OP_CODE(JMP));
		labelsArrayList = append(labelsArrayList, p.ip);
		p.genAddressInt(0);

		// Fill JFALSE hole
		save := p.ip;
		p.ip = hole;
		p.genAddressInt(save);

		p.ip = save;

		// PUSH the original eToken variable back to prepare for the next eql condition case label
		if (!p.currentToken.getTokenVal() == "TK_END"){
			symbol := p.symbolTable.lookup(eToken.getTokenVal());
			if (symbol != nil) {
				p.genOpCode(OP_CODE(PUSH));
				p.genAddressInt(symbol.getAddress());
			}
		}
	}

	p.match("TK_END");
	p.match("TK_SEMI_COLON");

	save := p.ip;

	// Fill all the labelHoles for JMP
	for labelHole := range  labelsArrayList{
		p.ip = labelHole;
		p.genAddressInt(save);
	}

	p.ip = save;
}

func (p *Parser) writeStat()  {
	p.match("TK_WRITELN");
	p.match("TK_OPEN_PARENTHESIS");
	var t PARSER_TYPE

	for (true) {
		symbol :=  p.symbolTable.lookup(p.currentToken.getTokenVal());



		if (symbol != nil) {
			if (symbol.getDataType() == PARSER_TYPE(A)) {
				// array
				p.currentToken.setTokenType("TK_AN_ARRAY");
				p.handleArrayAccess(symbol);

				p.genOpCode(OP_CODE(GET));

				t = symbol.getValueType();

			} else {
				// variable
				p.currentToken.setTokenType("TK_A_VAR");

				t = symbol.getDataType();
				p.genOpCode(OP_CODE(PUSH));
				p.genAddressInt(symbol.getAddress());
				p.match("TK_A_VAR");
			}
		} else {
			// literal
			t = p.getLitType(p.currentToken.getTokenType());
			//assert t != nil;
			switch (t) {
			case R:
				p.genOpCode(OP_CODE(PUSHF));
				p.genAddressFl(p.currentToken.getTokenVal());
				break;
			case I:
				p.genOpCode(OP_CODE(PUSHI));
				p.genAddressInt(p.currentToken.getTokenVal());
				break;
			case B:
				p.genOpCode(OP_CODE(PUSHI));
				if (p.currentToken.getTokenVal() == "true") {
					p.genAddressInt(1);
				} else {
					p.genAddressInt(0);
				}
				break;
			case C:
				p.genOpCode(OP_CODE(PUSHI));
				p.genAddressInt(p.currentToken.getTokenVal()[0]);
				break;
			}

			p.match(p.currentToken.getTokenType());
		}

		//assert t != null;
		switch (t) {
		case I:
			p.genOpCode(OP_CODE(PRINT_INT));
			break;
		case C:
			p.genOpCode(OP_CODE(PRINT_CHAR));
			break;
		case R:
			p.genOpCode(OP_CODE(PRINT_REAL));
			break;
		case B:
			p.genOpCode(OP_CODE(PRINT_BOOL));
			break;
		default:
			panic("Cannot write unknown type");

		}

		switch (p.currentToken.getTokenType()) {
		case "TK_COMMA":
			p.match("TK_COMMA");
			break;
		case "TK_CLOSE_PARENTHESIS":
			p.match("TK_CLOSE_PARENTHESIS");
			p.genOpCode(OP_CODE(PRINT_NEWLINE));
			return;
		default:
			panic("Current token type (" + p.currentToken.getTokenType() + ") is neither TK_COMMA nor TK_CLOSE_PARENTHESIS");
		}

	}
}

func (p *Parser) assignmentStat() {
	symbol := p.symbolTable.lookup(p.currentToken.getTokenVal());

	if (symbol != nil) {
		lhsType := symbol.getDataType();
		lhsAddress := symbol.getAddress();

		p.match("TK_A_VAR");

		p.match("TK_ASSIGNMENT");

		rhsType := p.E();
		if (lhsType == rhsType) {
			p.genOpCode(OP_CODE(POP));
			p.genAddressInt(lhsAddress);
		} else {
			panic("LHS type " + lhsType + " is not equal to RHS type: " + rhsType);
		}
	}
}


func (p *Parser) arrayAssignmentStat()  {
	symbol := p.symbolTable.lookup(p.currentToken.getTokenVal());
	if (symbol != nil) {

		p.handleArrayAccess(symbol);

		p.match("TK_ASSIGNMENT");


		rhsType := p.E();
		// Emit OP_CODE.PUT
		if (symbol.getValueType() == rhsType) {
			p.genOpCode(OP_CODE(PUT));
		}

	}
}


/*
    Condition
    C -> EC'
    C' -> < EC' | > EC' | <= EC' | >= EC' | = EC' | <> EC' | epsilon
     */
func (p* Parser) C() PARSER_TYPE  {
	e1 := p.E();
	for(p.currentToken.getTokenType() == "TK_LESS_THAN" ||
		p.currentToken.getTokenType() == "TK_GREATER_THAN" ||
		p.currentToken.getTokenType() == "TK_LESS_THAN_EQUAL" ||
		p.currentToken.getTokenType() == "TK_GREATER_THAN_EQUAL" ||
		p.currentToken.getTokenType() == "TK_EQUAL" ||
		p.currentToken.getTokenType() == "TK_NOT_EQUAL") {
		pred := p.currentToken.getTokenType();
		p.match(pred);
		e2 := p.T();

		e1 = p.emit(pred, e1, e2);
	}

	return e1;
}
/*
    Expression
    E -> TE'
    E' -> +TE' | -TE' | epsilon
     */
func (p *Parser) E() PARSER_TYPE {
	t1 := p.T()
	for (p.currentToken.getTokenType() == "TK_PLUS" || p.currentToken.getTokenType() == "TK_MINUS") {
		op := p.currentToken.getTokenType();
		p.match(op);
		t2 := p.T();

		t1 = p.emit(op, t1, t2);
	}

	return t1;
}


/*
    Term
    T -> FT'
    T' ->  *FT' | /FT' | epsilon
     */
func (p * Parser) T() PARSER_TYPE  {
	f1 := p.F()
	for(p.currentToken.getTokenType() == "TK_MULTIPLY" ||
		p.currentToken.getTokenType() == "TK_DIVIDE" ||
		p.currentToken.getTokenType() == ("TK_DIV")) {
		op := p.currentToken.getTokenType();
		p.match(op);
		f2 := p.F();

		f1 = p.emit(op, f1, f2);
	}
	return f1;
}


/*
    Factor
    F -> id | lit | (E) | not F | +F | -F
     */

func (p* Parser) F() PARSER_TYPE {
	switch (p.currentToken.getTokenType()) {
	case "TK_IDENTIFIER":
		symbol := p.symbolTable.lookup(p.currentToken.getTokenVal());
		if (symbol != nil) {
			if (symbol.getTokenType() == "TK_A_VAR") {
				// variable
				p.currentToken.setTokenType("TK_A_VAR");

				p.genOpCode(OP_CODE(PUSH));
				p.genAddressInt(symbol.getAddress());

				p.match("TK_A_VAR");
				return symbol.getDataType();
			} else if (symbol.getTokenType() == "TK_AN_ARRAY") {
				p.currentToken.setTokenType("TK_AN_ARRAY");

				p.handleArrayAccess(symbol);
				p.genOpCode(OP_CODE(GET));

				return symbol.getValueType();
			}
		} else {
			panic("Symbol not found (" + p.currentToken.getTokenVal() + ")");
		}
	case "TK_INTLIT":
		p.genOpCode(OP_CODE(PUSHI));
		p.genAddressInt(p.currentToken.getTokenVal());

		p.match("TK_INTLIT");
		return PARSER_TYPE(I);
	case "TK_FLOATLIT":
		p.genOpCode(OP_CODE(PUSHF));
		p.genAddressFl(p.currentToken.getTokenVal());

		p.match("TK_FLOATLIT");
		return PARSER_TYPE(R);
	case "TK_BOOLLIT":
		p.genOpCode(OP_CODE(PUSHI));

		s := p.currentToken.getTokenVal()
		if s != nil {
			p.genAddressInt(1)
		} else {
			p.genAddressInt(0)
		}

		p.match("TK_BOOLLIT");
		return PARSER_TYPE(B);
	case "TK_CHARLIT":
		p.genOpCode(OP_CODE(PUSHI));
		p.genAddressInt(p.currentToken.getTokenVal()[0]);

		p.match("TK_CHARLIT");
		return PARSER_TYPE(C);
	case "TK_STRLIT":
		i := 0
		for ( i < len(p.currentToken.getTokenType())) {
			p.genOpCode(OP_CODE(PUSHI));
			p.genAddressInt(p.currentToken.getTokenType()[i]);
			i++
		}

		p.match("TK_STRLIT");
		return PARSER_TYPE(S);
	case "TK_NOT":
		p.match("TK_NOT");
		return p.F();
	case "TK_OPEN_PARENTHESIS":
		p.match("TK_OPEN_PARENTHESIS");
		t := p.E();
		p.match("TK_CLOSE_PARENTHESIS");
		return t;
	default:
		panic("Unknown data type");
	}
	return 0
}

func (p *Parser) handleArrayAccess(symbol Symbol){
	p.match("TK_AN_ARRAY");
	p.match("TK_OPEN_SQUARE_BRACKET");



	varSymbol := p.symbolTable.lookup(p.currentToken.getTokenVal())
	if (varSymbol != nil) {
		t := varSymbol.getDataType();


		if (t != symbol.getIndexType()) {
			panic("Incompatible index type: (" +  t  +", "+ symbol.getIndexType()  + ")")
		}

		p.currentToken.setTokenType("TK_A_VAR");
		p.genOpCode(OP_CODE(PUSH));
		p.genAddressInt(varSymbol.getAddress());
		p.match("TK_A_VAR");

		p.match("TK_CLOSE_SQUARE_BRACKET");

		p.genOpCode(OP_CODE(PUSHI));

		switch (t) {
		case I:

			i1 := symbol.getLow();

			p.genAddressInt(i1);
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(SUB));

			// push element size
			p.genOpCode(OP_CODE(PUSHI));
			p.genAddressInt(4);

			p.genOpCode(OP_CODE(MULT));

			p.genOpCode(OP_CODE(PUSHI));
			p.genAddressInt(symbol.getAddress());

			p.genOpCode(OP_CODE(ADD));

			break;
		case C:
			c1 := symbol.getLow();

			p.genAddressInt(c1);
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(SUB));

			p.genOpCode(OP_CODE(PUSHI));
			p.genAddressInt(symbol.getAddress());

			p.genOpCode(OP_CODE(ADD));

			break;
		}
	} else {


		index := p.currentToken.getTokenVal();
		t := p.E();

		if (t != symbol.getIndexType()) {
			panic("Incompatible index type: ( " + t + " , " +  symbol.getIndexType() + ")");
		}

		p.match("TK_CLOSE_SQUARE_BRACKET");

		p.genOpCode(OP_CODE(PUSHI));

		switch (t) {
		case I:

			i1 := symbol.getLow();
			i2 := symbol.getHigh();

			// range check:
			if (index < i1 || index > i2) {
				panic("Index " + index + "is not within range " + i1 + "to " + i2);
			}

			p.genAddressInt(i1);
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(SUB));

			// push element size
			p.genOpCode(OP_CODE(PUSHI));
			p.genAddressInt(4);

			p.genOpCode(OP_CODE(MULT));

			p.genOpCode(OP_CODE(PUSHI));
			p.genAddressInt(symbol.getAddress());

			p.genOpCode(OP_CODE(ADD));

			break;
		case C:
			c1 := symbol.getLow();
			c2 := symbol.getHigh();

			// range check
			if (index[0] < c1 || index[0] > c2) {
				panic("Index " + index + "is not within range " + c1 + "to " + c2);
			}

			p.genAddressInt(c1);
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(SUB));

			p.genOpCode(OP_CODE(PUSHI));
			p.genAddressInt(symbol.getAddress());

			p.genOpCode(OP_CODE(ADD));

			break;
		}

	}
}


func (p* Parser) match(tokenType string) {
	if !tokenType == p.currentToken.getTokenType(){
		panic("Token type (%s) does not match current token type (%s)" + tokenType + " " + p.currentToken.getTokenType())
	} else {
		p.getToken()

	}
}

func (p *Parser) getToken () {
	if p.tokens.hasNext() {
		p.currentToken = p.tokens.getNext()
	}

}

func (p* Parser) getLitType(tokenType string) PARSER_TYPE{
	switch tokenType {
	case "TK_INTLIT":
		return PARSER_TYPE(I);
	case "TK_FLOATLIT":
		return PARSER_TYPE(R);
	case "TK_CHARLIT":
		return PARSER_TYPE(C);
	case "TK_BOOLLIT":
		return PARSER_TYPE(B);
	default:
		return nil;

	}
}

func (p* Parser) genAddressInt(add int32)  {
	buf := new (bytes.Buffer)
	var data = []interface{}{
		int32(add),
	}
	for _, v := range data {
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
	}
	for _, v := range buf.Bytes() {
		p.byteArray[p.ip] = v
		p.ip++
	}
}

func (p* Parser) genAddressFl(add float32)  {
	buf := new (bytes.Buffer)
	var data = []interface{}{
		float32(add),
	}
	for _, v := range data {
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
	}
	for _, v := range buf.Bytes() {
		p.byteArray[p.ip] = v
		p.ip++
	}
}



func (p* Parser) genOpCode(b OP_CODE)  {
	p.byteArray[p.ip] = (byte)(b);
	p.ip++
}



func (p *Parser) emitBool(pred OP_CODE, t1 PARSER_TYPE, t2 PARSER_TYPE ) PARSER_TYPE {
	if (t1 == t2) {
		p.genOpCode(pred);
		return PARSER_TYPE(B);
	} else if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(R)) {
		p.genOpCode(OP_CODE(XCHG));
		p.genOpCode(OP_CODE(CVR));
		p.genOpCode(pred);
		return PARSER_TYPE(B);
	} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(I)) {
		p.genOpCode(OP_CODE(CVR));
		p.genOpCode(pred);
		return PARSER_TYPE(B);
	}

	return nil;
}

func (p *Parser) emit (op string, t1 PARSER_TYPE, t2 PARSER_TYPE) PARSER_TYPE {
	switch op {
	case "TK_PLUS":
		if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(ADD));
			return PARSER_TYPE(I);
		} else if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FADD));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FADD));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(FADD));
			return PARSER_TYPE(R);
		}
	case "TK_MINUS":
		if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(SUB));
			return PARSER_TYPE(I);
		} else if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FSUB));
			return PARSER_TYPE(R)
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FSUB));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(FSUB));
			return PARSER_TYPE(R);
		}
	case "TK_MULTIPLY":
		if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(MULT));
			return PARSER_TYPE(I);
		} else if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FMULT));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FMULT));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(FMULT));
			return PARSER_TYPE(R);
		}
	case "TK_DIVIDE":
		if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(FDIV));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(XCHG));
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FDIV));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(CVR));
			p.genOpCode(OP_CODE(FDIV));
			return PARSER_TYPE(R);
		} else if (t1 == PARSER_TYPE(R) && t2 == PARSER_TYPE(R)) {
			p.genOpCode(OP_CODE(FDIV));
			return PARSER_TYPE(R);
		}
	case "TK_DIV":
		if (t1 == PARSER_TYPE(I) && t2 == PARSER_TYPE(I)) {
			p.genOpCode(OP_CODE(DIV));
			return PARSER_TYPE(I);
		}
	case "TK_LESS_THAN":
		return p.emitBool(OP_CODE(LSS), t1, t2);
	case "TK_GREATER_THAN":
		return p.emitBool(OP_CODE(GTR), t1, t2);
	case "TK_LESS_THAN_EQUAL":
		return p.emitBool(OP_CODE(LEQ), t1, t2);
	case "TK_GREATER_THAN_EQUAL":
		return p.emitBool(OP_CODE(GEQ), t1, t2);
	case "TK_EQUAL":
		return p.emitBool(OP_CODE(EQL), t1, t2);
	case "TK_NOT_EQUAL":
		return p.emitBool(OP_CODE(NEQL), t1, t2);
	}

	return nil;
}
