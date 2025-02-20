package main

import "fmt"

// func sum(arr []int) int {
// 	sum := 0
// 	for i := 0; i < len(arr); i++ {
// 		sum += arr[i]
// 	}
// 	return sum
// }

func sum(arr ...int) int {
	sum := 0
	for i := 0; i < len(arr); i++ {
		sum += arr[i]
	}
	return sum
}

func main() {
	// fmt.Println(sum(1, 2, 3))
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println(sum(nums...))
}
