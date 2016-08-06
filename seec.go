package main

import (
	"fmt"
	"os"

	"github.com/VonC/godbg"
	"github.com/VonC/godbg/exit"
)

var ex *exit.Exit
var pdbg *godbg.Pdbg

func init() {
	ex = exit.Default()
	if os.Getenv("dbg") != "" {
		pdbg = godbg.NewPdbg()
	} else {
		pdbg = godbg.NewPdbg(godbg.OptExcludes([]string{"/seec.go"}))
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:       go run seec.go <sha1>")
		fmt.Println("       dbg=1 go run seec.go <sha1> for debug information")
		fmt.Println(`       cmd /v /c "set dbg=1 && bin\seec* <sha1>" for debug information`)
		ex.Exit(0)
	}
	sha1 := os.Args[1]
	pdbg.Pdbgf("sha1: %s", sha1)
}
