package main

import "fmt"

/*
	可以利用多核的优势把一段粗粒度逻辑分解成多个 goroutine 执行
 */

func generator(max int) <-chan int{
	out := make(chan int, 100)
	go func() {
		for i := 1; i <= max; i++ {
			out <- i
		}
		close(out)
	}()
	return out
}

func power(in <-chan int) <-chan int{
	out := make(chan int, 100)
	go func() {
		for v := range in {
			out <- v * v
		}
		close(out)
	}()
	return out
}

func sum(in <-chan int) <-chan int{
	out := make(chan int, 100)
	go func() {
		var sum int
		for v := range in {
			sum += v
		}
		out <- sum
		close(out)
	}()
	return out
}

func main() {
	// [1, 2, 3]
	fmt.Println(<-sum(power(generator(3))))
}
