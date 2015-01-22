package main

import (
	"fmt"
	"log"
	"strconv"
)

//Do some works.
func doWork(i int, ch chan int) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("work wrong: ", err)
			ch <- i //Warning! Must tell channel as well.
		}
	}()
	if i%10 == 0 {
		panic("Some panic : " + strconv.Itoa(i))
	} else {
		fmt.Println("work fine: ", strconv.Itoa(i))
		ch <- i
	}
}

func main() {
	total := 100
	ch := make(chan int, total)
	for i := 0; i < total; i++ {
    //Doing asynchronously.
		go doWork(i, ch)
	}

	for i := 0; i < total; i++ {
		<-ch
	}
}
