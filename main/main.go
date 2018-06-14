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
	c := crawler.New(s, &receiver{})
	t, err := c.Task(0, "http://bbc.co.uk", ``)
	if err != nil {
		panic(err)
	}
	c.Run(t)
	time.Sleep(time.Second*30)
}

type receiver struct {}

func (r *receiver) Receive(url string, taskId int64, userId int64) {
	fmt.Println(url)
}