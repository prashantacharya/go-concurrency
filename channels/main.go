package main

type Result struct {
	Num    int
	Square int
}

func main() {
	c := make(chan Result, 10)

	totalTasks := 3

	for i := 0; i < totalTasks; i++ {
		go Task(i, c)
	}

	for i := 0; i < totalTasks; i++ {
		v := <-c
		println(v.Num, v.Square)
	}
}

func Task(i int, c chan Result) {
	c <- Result{i, Square(i)}
}

func Square(i int) int {
	return i * i
}
