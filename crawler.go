package crawler

import (
	"time"
	"regexp"
	"log"
	"io"
	"io/ioutil"
	"net/url"
)

type Record struct {
	TaskId int64
	Url string
}

type Task struct {
	id int64
	UserId int64
	Url string
	RegexpRule string
	regexpCompiled *regexp.Regexp

	UrlScheme string
	UrlHost string
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
	Receive(url string, taskId int64)
}

type Crawler struct {
	storage Storage
	receiver Receiver
	queue *HttpQueue
	tasks []*Task
}

func New(s Storage, receiver Receiver) (*Crawler, error) {
	/*
	tasks, err := s.ListTasks(0)
	if err != nil {
		return nil, err
	}
	*/

	c := &Crawler{
		storage: s,
		receiver: receiver,
	//	tasks: tasks,
	}
	c.queue = NewHttqQueue(2, c.rcv)
	return c, nil
}

func (c *Crawler) Task(UserId int64, Url, RegexpRule string) (int64, error) {
	u, err := url.Parse(Url)
	if err != nil {
		return -1, err
	}

	regexpCompiled, err := regexp.Compile(RegexpRule)
	if err != nil {
		return -1, err
	}

	t := &Task{
		UserId: UserId,
		Url: Url,
		RegexpRule: RegexpRule,
		regexpCompiled: regexpCompiled,
		UrlScheme: u.Scheme,
		UrlHost: u.Host,
	}

	t.id, err = c.storage.AddTask(t)
	c.tasks = append(c.tasks, t)
	return t.id, err
}

func (c *Crawler) GetTask(taskId int64) *Task {
	for _, t := range c.tasks {
		if t.id == taskId {
			return t
		}
	}
	return nil
}

func (c *Crawler) Run(taskId int64) {
	c.run(taskId)
}

func (c *Crawler) run(taskId int64) {
	task:= c.GetTask(taskId)
	if task != nil {
		c.processUrl(task.Url, taskId)
	}
}

func (c *Crawler) rcv(Url string, data io.ReadCloser, id int64, err error) {
	if err != nil {
		log.Printf("Failed to get url: %v. Error: %v", Url, err)
		return
	}
	defer data.Close()

	err = c.storage.AddRecord(&Record{
		TaskId: id,
		Url: Url,
	})
	if err != nil {
		log.Printf("Error saving record: %v", err)
		return
	}

	b, err := ioutil.ReadAll(data)
	if err != nil {
		log.Printf("Failed to read data. Url: %v. Error: %v", Url, err)
		return
	}
	res := hrefRegex.FindAllSubmatch(b, -1)
	urls := make([]string, len(res))
	for i, r := range res {
		urls[i] = string(r[1])
	}

	t := c.GetTask(id)

	for _, u := range urls {
		parsed, _ := url.Parse(u)
		if parsed.Host == "" {
			parsed.Host = t.UrlHost
		}
		if parsed.Scheme == "" {
			parsed.Scheme = t.UrlScheme
		}
		if t.regexpCompiled.MatchString(parsed.String()) {
			c.processUrl(parsed.String(), id)
		}
	}

	c.receiver.Receive(Url, id)
}

func (c *Crawler) processUrl(url string, taskId int64) {
	exists, _ := c.storage.Exists(url, nil)
	/*
	if err != nil {
		log.Printf("Error checking url existence: %v", err)
		return
	}
	*/

	if !exists {
		go c.queue.Put(url, taskId)
	}
}


var hrefRegex *regexp.Regexp = regexp.MustCompile(`\shref=[\'"]?([^\'" >]+)`)
