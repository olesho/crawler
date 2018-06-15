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
	Checked bool
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

func (t *Task) GetID() int64 {
	return t.id
}

func (t *Task) Compile() error {
	u, err := url.Parse(t.Url)
	if err != nil {
		return err
	}

	regexpCompiled, err := regexp.Compile(t.RegexpRule)
	if err != nil {
		return err
	}

	t.UrlHost = u.Host
	t.UrlScheme = u.Scheme
	t.regexpCompiled = regexpCompiled
	return nil
}

type Storage interface {
	GetTask(id int64) (*Task, error)
	AddTask(*Task) (int64, error)
	RemoveTask(id int64) error
	ListTasks(uid int64) ([]*Task, error)

	AddRecord(*Record) error
	SetRecordChecked(url string) error
	ListUncheckedRecords(taskId int64) ([]*Record, error)
	Exists(url string, maxTimeStamp *time.Time) (bool, error)
}

type Receiver interface {
	Receive(url string, taskId int64, data []byte)
}

type Crawler struct {
	storage Storage
	receiver Receiver
	queue *HttpQueue
	tasks []*Task
}

func New(s Storage, receiver Receiver) (*Crawler, error) {
	tasks, err := s.ListTasks(0)
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		err := t.Compile()
		if err != nil {
			return nil, err
		}
	}

	c := &Crawler{
		storage: s,
		receiver: receiver,
		tasks: tasks,
	}
	c.queue = NewHttqQueue(2, c.rcv)
	return c, nil
}

func (c *Crawler) Task(UserId int64, Url, RegexpRule string) (int64, error) {
	t := &Task{
		UserId: UserId,
		Url: Url,
		RegexpRule: RegexpRule,
	}
	t.Compile()

	var err error
	t.id, err = c.storage.AddTask(t)

	c.tasks = append(c.tasks, t)
	return t.id, err
}

func (c *Crawler) Tasks() []*Task {
	return c.tasks
}

func (c *Crawler) getTask(taskId int64) *Task {
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
	task:= c.getTask(taskId)
	if task != nil {
		c.processUrl(task.Url, taskId)
		recs, err := c.storage.ListUncheckedRecords(taskId)
		if err != nil {
			log.Printf("Unable to list unchecked records: %v", err)
		}
		for _, r := range recs {
			go c.queue.Put(r.Url, taskId)
		}
	}
}

func (c *Crawler) rcv(Url string, data io.ReadCloser, id int64, err error) {
	if err != nil {
		log.Printf("Failed to get url: %v. Error: %v", Url, err)
		return
	}
	defer data.Close()

	/*
	err = c.storage.AddRecord(&Record{
		TaskId: id,
		Url: Url,
		Checked: true,
	})
	*/
	err = c.storage.SetRecordChecked(Url)
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
	c.receiver.Receive(Url, id, b)

	t := c.getTask(id)
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
}

func (c *Crawler) processUrl(Url string, taskId int64) {
	exists, err := c.storage.Exists(Url, nil)
	if err != nil {
		log.Printf("Error checking url existence: %v", err)
		return
	}

	if !exists {
		err = c.storage.AddRecord(&Record{
			TaskId: taskId,
			Url: Url,
			Checked: false,
		})
		go c.queue.Put(Url, taskId)
	}
}


var hrefRegex *regexp.Regexp = regexp.MustCompile(`\shref=[\'"]?([^\'" >]+)`)
