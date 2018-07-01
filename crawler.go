package crawler

import (
	"regexp"
	"log"
	"net/url"
	"time"
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
	filter func(url string) bool

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
	GetRecord(url string, t *time.Time) (*Record, error)
}

type Receiver interface {
	Receive(url string, taskId int64, data []byte, err error) bool
}

type Crawler struct {
	storage Storage
	receiver Receiver
	queue *HttpQueue
	tasks []*Task
	done chan struct{}
}

func New(s Storage, receiver Receiver, queue *HttpQueue) (*Crawler, error) {
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
		done: make(chan struct{}),
	}
	c.queue = queue
	c.queue.SetReceiver(c)
	return c, nil
}

func (c *Crawler) Done() chan struct{} {
	return c.done
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

	t.filter = func(url string) bool {
		return t.regexpCompiled.MatchString(url)
	}

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
		if len(recs) == 0 {
			c.done <- struct{}{}
		}
		for _, r := range recs {
			go c.queue.Put(r.Url, taskId)
		}
	}
}

func (c *Crawler) SetFilterFunc(taskId int64, f func(url string) bool) {
	c.getTask(taskId).filter = f
}

func (c *Crawler) Receive(Url string, id int64, data []byte, requestErr error) bool {
	if requestErr != nil {
		log.Printf("Failed to get url: %v. Error: %v", Url, requestErr)
		return false
	}

	res := hrefRegex.FindAllSubmatch(data, -1)
	urls := make([]string, len(res))
	for i, r := range res {
		urls[i] = string(r[1])
	}

	t := c.getTask(id)
	for _, u := range urls {
		parsed, _ := url.Parse(u)
		if parsed.Host == "" {
			parsed.Host = t.UrlHost
		}
		if parsed.Scheme == "" {
			parsed.Scheme = t.UrlScheme
		}
		//if t.regexpCompiled.MatchString(parsed.String()) {
			if t.filter(parsed.String()) {
				c.processUrl(parsed.String(), id)
			}

		//}
	}
	if c.receiver.Receive(Url, id, data, requestErr) {
		err := c.storage.SetRecordChecked(Url)
		if err != nil {
			log.Printf("Error saving record: %v", err)
			return false
		}
	}
	if c.queue.Done() {
		c.done <- struct{}{}
	}
	return true
}

func (c *Crawler) processUrl(Url string, taskId int64) {
	r, _:= c.storage.GetRecord(Url, nil)

	if r != nil {
		if r.Checked {
			return
		}
		go c.queue.Put(Url, taskId)
		return
	}

	if r == nil {
		err := c.storage.AddRecord(&Record{
			TaskId: taskId,
			Url: Url,
			Checked: false,
		})
		if err != nil {
			log.Printf("Unable to add record: %v", err)
		}
	}
	go c.queue.Put(Url, taskId)
}


var hrefRegex *regexp.Regexp = regexp.MustCompile(`\shref=[\'"]?([^\'">]+)`)
