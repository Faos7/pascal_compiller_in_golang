package main

import "strconv"
import (
	"math"
	"fmt"
	"os"
	"strings"
	"encoding/binary"
	"bytes"
)

type Simulator struct {
	dp 		int
	ip 		int
	dataArray 	[]byte
	instructions 	[1000]byte
//	stack
	stack		Stack
}

func NewSimulator (intsructions []byte) *Simulator  {
	return &Simulator{
		dp 	 :		 0,
		ip 	 :	         0,
		stack	 : 	NewStack(),
	}
}

func (s * Simulator) simulate()  {
	var opCode OP_CODE

	for opCode != OP_CODE(HALT) {
		opCode = s.getOpCode();
		switch (opCode) {
		case PUSH:
			s.push();
		case PUSHI:
			s.pushi();
		case PUSHF:
			s.pushf();
		case POP:
			s.pop();
		case GET:
			s.get();
		case PUT:
			s.put();
		case CVR:
			s.cvr();
		case XCHG:
			s.xchg();
		case JMP:
			s.jmp();
		case PRINT_REAL:
			s.printReal();
		case PRINT_INT:
			s.printInt();
		case PRINT_BOOL:
			s.printBool();
		case PRINT_CHAR:
			s.printChar();
		case PRINT_NEWLINE:
			fmt.Println()
		case HALT:
			s.halt();
		case EQL:
			s.eql();
		case NEQL:
			s.neql();
		case LSS:
			s.less();
		case LEQ:
			s.lessEql();
		case GTR:
			s.greater();
		case GEQ:
			s.greaterEql();
		case JFALSE:
			s.jfalse();
		case JTRUE:
			s.jtrue();
		case ADD:
			s.add();
		case FADD:
			s.fadd();
		case SUB:
			s.sub();
		case FSUB:
			s.fsub();
		case MULT:
			s.mult();
		case FMULT:
			s.fmult();
		case DIV:
			s.div();
		case FDIV:
			s.fdiv();
		default:
			panic("Unhandled case: " + opCode);
		}

	}
}

func (s * Simulator) put() string  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v2, _:= strconv.Atoi(val2)
	s.dp = v2

	if strings.Contains(val1, ".") {
		v1, _ := strconv.ParseFloat(val2, 32)
		buf := new (bytes.Buffer)
		var data = []interface{}{
			float32(v1),
		}
		for _, v := range data {
			err := binary.Write(buf, binary.LittleEndian, v)
			if err != nil {
				fmt.Println("binary.Write failed:", err)
			}
		}
		for _, v := range buf.Bytes() {
			s.dataArray[s.dp] = v
			s.dp++
		}
	} else {
		v1, _:= strconv.Atoi(val2)
		buf := new (bytes.Buffer)
		var data = []interface{}{
			int32(v1),
		}
		for _, v := range data {
			err := binary.Write(buf, binary.LittleEndian, v)
			if err != nil {
				fmt.Println("binary.Write failed:", err)
			}
		}
		for _, v := range buf.Bytes() {
			s.dataArray[s.dp] = v
			s.dp++
		}
	}

	return val1
}

func (s * Simulator) jtrue()  {
	val1 := s.stack.Pop().Value
	if val1 == "true" {
		s.ip = s.getAddressVal()
	} else {
		s.getAddressVal()
	}
}

func (s * Simulator) jfalse()  {
	val1 := s.stack.Pop().Value
	if val1 == "false" {
		s.ip = s.getAddressVal()
	} else {
		s.getAddressVal()
	}
}

func (s * Simulator) eql()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	v3 := v1 == v2
	if v3 {
		s.stack.Push(NewNode("true"))
	} else {
		s.stack.Push(NewNode("false"))
	}

}

func (s * Simulator) neql()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	v3 := v1 != v2
	if v3 {
		s.stack.Push(NewNode("true"))
	} else {
		s.stack.Push(NewNode("false"))
	}
}

func (s * Simulator) less()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	v3 := v1 < v2
	if v3 {
		s.stack.Push(NewNode("true"))
	} else {
		s.stack.Push(NewNode("false"))
	}
}

func (s * Simulator) get()  {
	val1 := s.stack.Pop().Value
	v1, _:= strconv.Atoi(val1)
	s.dp = v1
	s.stack.Push(NewNode(s.getData(s.dp)))
}

func (s * Simulator) greater()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	v3 := v1 > v2
	if v3 {
		s.stack.Push(NewNode("true"))
	} else {
		s.stack.Push(NewNode("false"))
	}
}

func (s * Simulator) lessEql()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	v3 := v1 <= v2
	if v3 {
		s.stack.Push(NewNode("true"))
	} else {
		s.stack.Push(NewNode("false"))
	}
}

func (s * Simulator) greaterEql()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	v3 := v1 >= v2
	if v3 {
		s.stack.Push(NewNode("true"))
	} else {
		s.stack.Push(NewNode("false"))
	}
}

func (s * Simulator) printReal()  {
	v1 := s.stack.Pop().Value
	if strings.Contains(v1, ".") {
		fmt.Println(v1)
	} else {
		v2, _ := strconv.Atoi(v1)
		buf := new (bytes.Buffer)
		var byteArray [4]byte
		var data = []interface{}{
			int32(v2),
		}
		for _, v := range data {
			err := binary.Write(buf, binary.LittleEndian, v)
			if err != nil {
				fmt.Println("binary.Write failed:", err)
			}
		}
		i :=0
		for _, v := range buf.Bytes() {
			byteArray[i] = v
			i++
		}

		f := math.Float32frombits(byteArray)

		fmt.Println(f)

	}
}

func (s * Simulator) printBool()  {
	v1 := s.stack.Pop().Value
	v2, _ := strconv.Atoi(v1)
	if v2 == 1 {
		fmt.Println("True")
	} else {
		fmt.Println("False")
	}
}

func (s * Simulator) printInt()  {
	fmt.Println(s.stack.Pop().Value)
}

func (s * Simulator) printChar()  {
	v1 := s.stack.Pop().Value
	v2, _ := strconv.Atoi(v1)
	c := rune(v2)
	fmt.Println(c)
}

func (s * Simulator) add()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _:= strconv.Atoi(val1)
	v2, _:= strconv.Atoi(val2)
	var v int32 = v1 + v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator) fadd() {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	var v int32 = v1 + v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator) sub()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _:= strconv.Atoi(val1)
	v2, _:= strconv.Atoi(val2)
	var v int32 = v1 - v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator) fsub()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	var v int32 = v1 - v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator) mult() {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _:= strconv.Atoi(val1)
	v2, _:= strconv.Atoi(val2)
	var v int32 = v1 * v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator) fmult() {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	var v int32 = v1 * v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator) fdiv()  {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _ := strconv.ParseFloat(val1, 32)
	v2, _ := strconv.ParseFloat(val2, 32)
	var v int32 = v1 / v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator) div() {
	val1 := s.stack.Pop().Value;
	val2 := s.stack.Pop().Value;
	v1, _:= strconv.Atoi(val1)
	v2, _:= strconv.Atoi(val2)
	var v int32 = v1 / v2
	s.stack.Push(NewNode(v))
}

func (s * Simulator)cvr()  {
	val := s.stack.Pop().Value
	i, err := strconv.ParseFloat(val, 32)
	if err == nil {
		s.stack.Push(NewNode(i))
	}
}

func (s * Simulator)xchg()  {
	val1 := s.stack.Pop();
	val2 := s.stack.Pop()
	s.stack.Push(val1)
	s.stack.Push(val2)
}

func (s *Simulator)pushf()  {
	val := s.getFloatVal()
	s.stack.Push(NewNode(val))
}

func (s *Simulator)pushi()  {
	val := s.getAddressVal()
	s.stack.Push(NewNode(val))
}

func (s *Simulator) push()  {
	s.dp = s.getAddressVal()
	s.stack.Push(NewNode(s.getData(s.dp)))
}

func (s * Simulator) pop() string {
	buf:= new (bytes.Buffer)
	val := s.stack.Pop().Value
	s.dp = s.getAddressVal()

	//var valBytes []byte;

	if strings.Contains(val, ".") {
		i, err :=strconv.ParseFloat(val, 32)
		var data = []interface{}{
			float32(i),
		}
		if err == nil {
			for _, v := range data {
				err := binary.Write(buf, binary.LittleEndian, v)
				if err != nil {
					fmt.Println("binary.Write failed:", err)
				}
			}
		}
	} else {
		i, err :=strconv.Atoi(val)
		var data = []interface{}{
			int32(i),
		}
		if err == nil {
			for _, v := range data {
				err := binary.Write(buf, binary.LittleEndian, v)
				if err != nil {
					fmt.Println("binary.Write failed:", err)
				}
			}
		}
	}

	for _, v := range buf.Bytes() {
		s.dataArray[s.dp] = v
		s.dp++
	}

	return val
}

func (s *Simulator) jmp() {
	s.ip = s.getAddressVal()
}

func (s *Simulator)halt () {
	fmt.Println("\nProgram finished with exit code 0\n");
	os.Exit(0);
}

func (s * Simulator) getAddressVal() int32 {
	var valArray [4]byte
	for i:=0; i < 4; i++ {
		valArray[i] = s.dataArray[s.ip]
		s.ip++
	}
	x, _ := strconv.Atoi(string(valArray))
	return x
}

func (s * Simulator) getFloatVal() float32 {
	var valArray [4]byte
	for i:=0; i < 4; i++ {
		valArray[i] = s.dataArray[s.ip]
		s.ip++
	}
	x := math.Float32frombits(valArray)
	return x
}

func (s *Simulator) getData(dp int) int32 {
var valArray [4]byte
	for i:=0; i < 4; i++ {
		valArray[i] = s.dataArray[dp]
		dp++
	}
	x, _ := strconv.Atoi(string(valArray))
	return x
}

func (s *Simulator) getOpCode () OP_CODE  {
	op := s.instructions[s.ip]
	s.ip++
	return OP_CODE(op)
}
