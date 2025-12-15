package main

import (
	"LinkMaskirator/service"
)

func main() {

	producer := service.NewFileProducer("C:/GoLand/GoCourse/LinkMaskirator/txtFiles/links.txt")

	presenter := service.NewFilePresenter("")

	ser := service.NewService(producer, presenter)

	ser.Run()
}
