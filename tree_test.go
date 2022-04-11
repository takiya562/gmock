package gmock

import (
	"testing"
)

type testRequests []struct {
	path string
	res  bool
}

func check(t *testing.T, tree *node, requests testRequests) {
	for _, request := range requests {
		value := tree.getValue(request.path)
		if value != request.res {
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
		{"/a", true},
		{"/", false},
		{"/hi", true},
		{"/contact", true},
		{"/co", true},
		{"/con", false},
		{"cona", false},
		{"no", false},
		{"/ab", true},
		{"/α", true},
		{"/β", true},
	})
}
