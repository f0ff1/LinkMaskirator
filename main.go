package main

import (
	"fmt"
	"time"

	"LinkMaskirator/service"

)

func main() {
	start := time.Now()

	producer := service.NewFileProducer("C:/GoLand/GoCourse/LinkMaskirator/txtFiles/links.txt")

	presenter := service.NewFilePresenter("")

	ser := service.NewService(producer, presenter)

	ser.Run()

	elapsed := time.Since(start)
	fmt.Printf("Программа выполнилась за %v\n", elapsed)
	fmt.Printf("В миллисекундах: %d ms\n", elapsed.Milliseconds())
	fmt.Printf("В микросекундах: %d µs\n", elapsed.Microseconds())
	fmt.Printf("В наносекундах: %d ns\n", elapsed.Nanoseconds())
}
