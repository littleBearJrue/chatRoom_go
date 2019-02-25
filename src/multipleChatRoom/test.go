package main

import "fmt"

func main() {
	string := "@"
	switch string {
	case "@":
		fmt.Println("@")
	default:
		fmt.Println("default!!!")
	}
}
