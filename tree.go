package gmock

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/takiya562/go-mock/internal/bytesconv"
)

var resPath = "responses/"

type node struct {
	path     string
	indices  string
	priority uint32
	children []*node
	fullPath string
	leaf     bool
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		cs[newPos], cs[newPos-1] = cs[newPos-1], cs[newPos]
	}

	if newPos != pos {
		n.indices = n.indices[:newPos] +
			n.indices[pos:pos+1] +
			n.indices[newPos:pos] + n.indices[pos+1:]
	}

	return newPos
}

func (n *node) addRoute(path string, response []byte) {
	fullPath := path
	n.priority++

	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		n.path = path
		n.fullPath = fullPath
		n.leaf = true
		return
	}

	parentFullPathIndex := 0

walk:
	for {
		i := longestCommonPrefix(path, n.path)

		if i < len(n.path) {
			child := node{
				path:     n.path[i:],
				children: n.children,
				indices:  n.indices,
				priority: n.priority - 1,
				fullPath: n.fullPath,
				leaf:     n.leaf,
			}

			n.children = []*node{&child}
			n.indices = bytesconv.BytesToString([]byte{n.path[i]})
			n.path = path[:i]
			n.leaf = false
			n.fullPath = fullPath[:parentFullPathIndex+i]
		}

		if i < len(path) {
			path = path[i:]
			c := path[0]

			for i, max := 0, len(n.indices); i < max; i++ {
				if c == n.indices[i] {
					parentFullPathIndex += len(n.path)
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			n.indices += bytesconv.BytesToString([]byte{c})
			child := &node{
				fullPath: fullPath,
				path:     path,
				leaf:     true,
			}
			n.children = append(n.children, child)
			n.incrementChildPrio(len(n.indices) - 1)
			n = child

			saveResponse(n.fullPath, response)
			return
		}

		if n.leaf {
			panic("'" + fullPath + "' is alredy registered")
		}

		n.fullPath = fullPath
		n.leaf = true
		saveResponse(n.fullPath, response)
		return
	}
}

func saveResponse(fullPath string, response []byte) {
	filename := fmt.Sprintf("%x", md5.Sum(bytesconv.StringToBytes(fullPath)))
	fp := filepath.Join(resPath, filename)
	f, err := os.Create(fp)
	if err != nil {
		log.Fatalf("Failed to create response file '%s': %s", fullPath, err)
	}
	defer f.Close()
	ioutil.WriteFile(fp, response, 0644)
}

type nodeValue struct {
	response []byte
	fullPath string
}

func (n *node) getValue(path string) (value nodeValue) {
walk:
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						n = n.children[i]
						continue walk
					}
				}
			}

			return nodeValue{
				fullPath: "",
			}
		}

		if path == prefix && n.leaf {
			return nodeValue{
				fullPath: n.fullPath,
				response: nil,
			}
		}

		return nodeValue{
			fullPath: "",
		}
	}
}
