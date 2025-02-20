package main

import "fmt"

func someFunction() ([]int, error) {
	return []int{1, 2, 3}, nil
}

func main() {
	result, _ := someFunction()
	fmt.Println(result)
}
