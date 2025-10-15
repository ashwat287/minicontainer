package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: minicontainer run <command>")
		return
	}

	switch os.Args[1] {
	case "run":
		Run()
	case "child":
		Child()
	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
