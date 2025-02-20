package main

import "fmt"

// Cấp phát bộ nhớ mà không phải khởi tạo điều gì

type Counter struct {
	count int
}

func (c *Counter) Increment() {
	c.count += 1
}

func NewCounter() *Counter {
	return new(Counter)
}

func main() {
	c := NewCounter()
	c.Increment()
	fmt.Println("Count: ", c.count)
}
