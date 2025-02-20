package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
}

type Rectangle struct {
	width, height float64
}

type Circle struct {
	radius float64
}

func (c *Circle) Area() float64 {
	return c.radius * c.radius * math.Pi
}
func (r *Rectangle) Area() float64 {
	return r.width * r.height
}

func calculateArea(s Shape) float64 {
	return s.Area()
}

func main() {
	rect := Rectangle{width: 5, height: 4}
	cir := Circle{radius: 2}
	fmt.Println(calculateArea(&rect))
	fmt.Println(calculateArea(&cir))
}
