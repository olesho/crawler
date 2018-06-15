package crawler
/*
import (
	"testing"
	"io"
	"fmt"
	"time"
	"io/ioutil"
)

func TestQueue(t *testing.T) {
	q := NewHttqQueue(3, func(url string, data io.ReadCloser, id int64, err error){
		if err != nil {
			fmt.Printf("Error getting: %v", err)
			return
		}
		defer data.Close()
		b, _ := ioutil.ReadAll(data)
		fmt.Printf("Url: %v\n Data: %v\n", url, string(b[:30]))
	})

	q.Put("https://bbc.co.uk", 1)
	q.Put("https://google.com", 2)
	q.Put("https://fb.com", 3)
	q.Put("https://yahoo.com", 4)
	q.Put("https://twitter.com", 5)
	q.Put("https://pravda.com.ua", 6)

	time.Sleep(time.Second * 10)
}
*/