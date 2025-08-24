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
			fmt.Println(err.Error())
		}

	case "clear":
		if len(args) != 1 {
			fmt.Println("Incorrect operands.")
			return
		}
		if err := Clear("."); err != nil {
			fmt.Println(err.Error())
		}

	case "add":
		if len(args) != 2 {
			fmt.Println("Incorrect operands.")
			return
		}
		if err := Add(".", args[1]); err != nil {
			fmt.Println(err.Error())
		}
	
	case "commit":
		if len(args) != 2 {
			fmt.Println("Incorrect operands.")
			return
		}
		if err := CommitCmd(".", args[1]); err != nil {
			fmt.Println(err.Error())
		}

	case "log":
		if len(args) != 1 {
			fmt.Println("Incorrect operands.")
			return
		}
		if err := LogCmd("."); err != nil {
			fmt.Println(err.Error())
		}

	case "checkout":
		// checkout -- <file>
		if len(args) == 3 && args[1] == "--" {
			if err := CheckoutHeadFile(".", args[2]); err != nil { fmt.Println(err.Error()) }
			return
		}
		// checkout <commit> -- <file>
		if len(args) == 4 && args[2] == "--" {
			if err := CheckoutCommitFile(".", args[1], args[3]); err != nil { fmt.Println(err.Error()) }
			return
		}
		// checkout <branch>
		if len(args) == 2 {
			if err := CheckoutBranchCmd(".", args[1]); err != nil { fmt.Println(err.Error()) }
			return
		}
		fmt.Println("Incorrect operands.")

	
	case "status":
		if len(args) != 1 {
			fmt.Println("Incorrect operands.")
			return
		}
		if err := StatusCmd("."); err != nil { fmt.Println(err.Error()) }

	case "global-log":
		if len(args) != 1 { fmt.Println("Incorrect operands."); return }
		if err := GlobalLogCmd("."); err != nil { fmt.Println(err.Error()) }

	case "find":
		if len(args) != 2 { fmt.Println("Incorrect operands."); return }
		if err := FindCmd(".", args[1]); err != nil { fmt.Println(err.Error()) }

	case "rm":
		if len(args) != 2 { fmt.Println("Incorrect operands."); return }
		if err := RmCmd(".", args[1]); err != nil { fmt.Println(err.Error()) }

	case "branch":
		if len(args) != 2 { fmt.Println("Incorrect operands."); return }
		if err := BranchCmd(".", args[1]); err != nil { fmt.Println(err.Error()) }
	
	case "rm-branch":
		if len(args) != 2 { fmt.Println("Incorrect operands."); return }
		if err := RmBranchCmd(".", args[1]); err != nil { fmt.Println(err.Error()) }

	case "reset":
		if len(args) != 2 { fmt.Println("Incorrect operands."); return }
		if err := ResetCmd(".", args[1]); err != nil { fmt.Println(err.Error()) }


	default:
		fmt.Println("No command with that name exists.")
	}
}
