package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	filename := os.Args[1]
	fmt.Println(filename)
	file_ext := filepath.Ext(filename)
	fmt.Println(file_ext)
	file_stat, err := os.Stat(filename)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(file_stat)
}
