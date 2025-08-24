// cmd/gitlet/main.go

package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please enter a command.")
		return
	}
	switch args[0] {
	case "init":
		if len(args) != 1 {
			fmt.Println("Incorrect operands.")
			return
		}
		if err := Init("."); err != nil {
			// spec wants exact messages printed and then stop.
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("No command with that name exists.")
	}
}
