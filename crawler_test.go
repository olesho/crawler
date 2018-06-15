package crawler

import (
	"testing"
	"net/url"
	"fmt"
)

/*
func TestProcess(t *testing.T) {
	a := assert.New(t)
	urls, err := process("http://bbc.co.uk")
	a.NoError(err, "Page not being processed")
	fmt.Println(urls)
}
*/

func TestUrl(t *testing.T) {
	u, _ := url.Parse("//secure-us.imrworldwide.com/")
	fmt.Printf("Scheme: %v Host: %v Full: %v", u.Scheme, u.Host, u.String())
	u.Scheme = "https"
	//u.Host = "bbc.co.ua"
	fmt.Printf("Scheme: %v Host: %v Full: %v", u.Scheme, u.Host, u.String())
}