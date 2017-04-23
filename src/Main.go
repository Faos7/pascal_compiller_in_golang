package main

import (
	"os"
	"io/ioutil"
	"fmt"
)

func main() {
	fileName := os.Args[1]
	code, err :=ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(-1)
	}

	s := string(code)
	tokenScanner := NewTokenScanner()
	tokenList := tokenScanner.scan(s)
	parser := NewParser(tokenList)
	parser.initSTRING_TYPE_MAP()

	var instructions []byte = parser.parse()
	simulator := NewSimulator(instructions)
	simulator.simulate()




}