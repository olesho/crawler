package main

import (
	"github.com/olesho/crawler"
	"os"
	"fmt"
	"time"
)

func main() {
	s, err := crawler.NewMysqlStorage(&crawler.MysqlConfig{
		User: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
	})
	if err != nil {
		panic(err)
	}

	c, err := crawler.New(s, &receiver{})
	if err != nil {
		panic(err)
	}

	availableTasks := c.Tasks()
	var taskId int64
	if len(availableTasks) > 0 {
		taskId = availableTasks[0].GetID()
	} else {
		taskId, err = c.Task(0, "https://bbc.co.uk/news/", `^https:\/\/bbc.co.uk\/news\/`)
		if err != nil {
			panic(err)
		}
	}

	c.Run(taskId)
	time.Sleep(time.Second*30)
}

type receiver struct {}

func (r *receiver) Receive(url string, taskId int64, data []byte) {
	fmt.Println(url, string(data[:100]))
}