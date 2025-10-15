package main

import (
	"fmt"
	"os"
	"strconv"
)

func must(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		panic(err)
	}
}

func writeFile(path, data string) {
	fmt.Printf("[debug] writing '%s' to %s\n", data, path)
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		fmt.Printf("Failed to write to %s: %v\n", path, err)
		panic(err)
	}
}

func getPid() string {
	return strconv.Itoa(os.Getpid())
}
