package main

import "fmt"

type aa struct {
	name string
}


func main()  {
	data := []string{"111", "222", "333", "444"}
	curData := append(data, "5555")
	data = curData
	fmt.Println(data)
	fmt.Println(curData)
}

