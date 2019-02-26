package main

import (
	"fmt"
	"strconv"
)

func main() {
	for i := 0; i < 5; i++ {
		fmt.Println("index_int: ", i)
		fmt.Println("index_str: ",string(i) )
		fmt.Println("index_str: ",strconv.Itoa(i) )
	}
}
