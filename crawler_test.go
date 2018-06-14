package crawler

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	a := assert.New(t)
	urls, err := process("http://bbc.co.uk")
	a.NoError(err, "Page not being processed")
	fmt.Println(urls)
}