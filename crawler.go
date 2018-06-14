package crawler

import (
	"time"
	"net/http"
	"regexp"
	"io/ioutil"
	"log"
	"fmt"
)

type Record struct {
	TaskId int64
	Url string
}

type Task struct {
	UserId int64
	Url string
	RegexpRule string
}

type Storage interface {
	GetTask(id int64) (*Task, error)
	AddTask(*Task) (int64, error)
	RemoveTask(id int64) error
	ListTasks(uid int64) ([]*Task, error)

	AddRecord(*Record) error
	ListRecords(taskId int64) ([]*Record, error)
	Exists(url string, maxTimeStamp *time.Time) (bool, error)
}

type Receiver interface {
	Receive(url string, taskId int64, userId int64)
}

type Crawler struct {
	storage Storage
	receiver Receiver
}

func New(s Storage, receiver Receiver) *Crawler {
	return &Crawler{
		storage: s,

	}
}

func (c *Crawler) Task(UserId int64, Url, RegexpRule string) (taskId int64, err error) {
	t := &Task{
		UserId: UserId,
		Url: Url,
		RegexpRule: RegexpRule,
	}

	return c.storage.AddTask(t)
}

func (c *Crawler) Run(taskId int64) {
	go c.run(taskId)
}

func (c *Crawler) run(taskId int64) {
	task, err := c.storage.GetTask(taskId)
	if err != nil {
		log.Printf("Error getting task: %v", err)
		return
	}
	if task != nil {
		c.processUrl(task.Url, taskId, task.UserId)
	}
}

func (c *Crawler) processUrl(url string, taskId, userId int64) {
	exists, err := c.storage.Exists(url, nil)
	if err != nil {
		log.Printf("Error checking url existence: %v", err)
	}

	if !exists {
		urls, err := process(url)
		if err != nil {
			log.Printf("Error processing: %v", err)
		}

		for _, u := range urls {
			c.processUrl(u, taskId, userId)
		}
	}
}

func (c *Crawler) receive(url string, taskId, userId int64) {
	err := c.storage.AddRecord(&Record{
		TaskId: taskId,
		Url: url,
	})
	if err != nil {
		log.Printf("Error saving record: %v", err)
		return
	}
	c.receiver.Receive(url, taskId, userId)
}

var hrefRegex *regexp.Regexp = regexp.MustCompile(`\shref=[\'"]?([^\'" >]+)`)

func process(url string) (urls []string, err error) {
	fmt.Printf("Getting %v ... ", url)
	r, err := http.Get(url)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	res := hrefRegex.FindAllSubmatch(b, -1)
	urls = make([]string, len(res))
	for i, r := range res {
		urls[i] = string(r[1])
	}
	fmt.Println("done")
	fmt.Println(urls)
	return
}
