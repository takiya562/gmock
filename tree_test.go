package gmock

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/takiya562/go-mock/internal/bytesconv"
)

type testRequests []struct {
	path  string
	route string
}

func check(t *testing.T, tree *node, requests testRequests) {
	for _, request := range requests {
		value := tree.getValue(request.path)
		if value.fullPath != request.route {
			t.Errorf("path mismatch for route '%s'", request.path)
		}
	}
}

func TestTreeAddAndGet(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/contact",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/α",
		"/β",
	}

	for _, route := range routes {
		tree.addRoute(route, []byte{})
	}

	check(t, tree, testRequests{
		{"/a", "/a"},
		{"/", ""},
		{"/hi", "/hi"},
		{"/contact", "/contact"},
		{"/co", "/co"},
		{"/con", ""},
		{"cona", ""},
		{"no", ""},
		{"/ab", "/ab"},
		{"/α", "/α"},
		{"/β", "/β"},
	})
}

func TestSaveResponse(t *testing.T) {
	saveResponse("/hi", []byte("hi"))
	filename := fmt.Sprintf("%x", md5.Sum(bytesconv.StringToBytes("/hi")))
	bs, err := ioutil.ReadFile("responses/" + filename)
	if err != nil {
		t.Errorf("Failed to read response file: %s", err)
	}
	if string(bs) != "hi" {
		t.Errorf("Response mismatch: %s", string(bs))
	}
}
