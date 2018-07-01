package crawler

import (
	"net/http"
	"github.com/pkg/errors"
	"io/ioutil"
	"compress/gzip"
	"io"
	"sync/atomic"
)

type HttpQueue struct {
	data chan request
	receiver Receiver
	headers map[string]string
	processed int32
}

type request struct {
	url string
	id int64
}

func NewHttqQueue(maxConns int, headers map[string]string) *HttpQueue {//, receiver func(url string, data io.ReadCloser, id int64, err error)) *HttpQueue {
	q := &HttpQueue{
		data: make(chan request),
		headers: headers,
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

func (q* HttpQueue) SetReceiver(r Receiver) {
	q.receiver = r
}

func (q* HttpQueue) Done() bool{
	return len(q.data) == 0 && q.processed == 0
}

func (q *HttpQueue) get(r request) {
	atomic.AddInt32(&q.processed, 1)
	q.processed++

	netTransport := &http.Transport{
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: netTransport,
	}

	req, err := http.NewRequest("GET", r.url, nil)
	if err != nil {
		q.receiver.Receive(r.url, r.id, nil, err)
	}

	for k, v := range q.headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		q.receiver.Receive(r.url, r.id, nil, err)
		atomic.AddInt32(&q.processed, -1)
		return
	}

	if resp != nil {
		defer resp.Body.Close()

		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			defer reader.Close()
		default:
			reader = resp.Body
		}

		b, _ := ioutil.ReadAll(reader)
		q.receiver.Receive(r.url, r.id, b,nil)
		atomic.AddInt32(&q.processed, -1)
		return
	}
	q.receiver.Receive(r.url, r.id, nil, errors.New("Unable to finish request"))
	atomic.AddInt32(&q.processed, -1)
}

