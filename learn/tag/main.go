package main

import "fmt"

func main() {
Outerloop:
	for j := range 3 {
		fmt.Println(j)
		for i := range 3 {
			if i%2 == 1 {
				break Outerloop
			}
			fmt.Println(i)
		}
	}
}
