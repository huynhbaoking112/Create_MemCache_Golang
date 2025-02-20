package main

import "fmt"

const (
	_ = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)
const (
	Readable   = 1 << iota // 1 << 0 = 001
	Writable               // 1 << 1 = 010
	Executable             // 1 << 2 = 100
)

func main() {
	// fmt.Println(Monday)
	// fmt.Println(Tuesday)
	// fmt.Println(Wednesday)
	// fmt.Println(Thursday)
	// fmt.Println(Friday)
	// fmt.Println(Saturday)
	// fmt.Println(Sunday)
	fmt.Println(Readable)
	fmt.Println(Writable)
	fmt.Println(Executable)

}
