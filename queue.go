package crawler

import (
	"net/http"
	"io/ioutil"
	"io"
	"github.com/pkg/errors"
)

type HttpQueue struct {
	data chan request
	receiver func(url string, data io.ReadCloser, id int64, err error)
}

type request struct {
	url string
	id int64
}

func NewHttqQueue(maxConns int, receiver func(url string, data io.ReadCloser, id int64, err error)) *HttpQueue {
	q := &HttpQueue{
		data: make(chan request),
		receiver: receiver,
	}

	for i := 0; i < maxConns; i++ {
		go func() {
			for  {
				q.get(<- q.data)
			}
		}()
	}
	return q
}

func (q *HttpQueue) Put(url string, id int64) {
	q.data <- request{url, id}
}

func (q *HttpQueue) get(r request) {
	resp, err := http.Get(r.url)
	if err != nil {
		q.receiver(r.url, nil, r.id, err)
		return
	}
	if resp != nil {
		q.receiver(r.url, resp.Body, r.id, nil)
		return
	}
	q.receiver(r.url, nil, r.id, errors.New("Unable to finish request"))
}

func Get(url string) (urls []string, err error) {
	//fmt.Printf("Getting %v ... ", url)
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
	//fmt.Println("done")
	//fmt.Println(urls)
	return
}
