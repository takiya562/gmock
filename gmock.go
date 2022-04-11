package gmock

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/takiya562/gmock/internal/bytesconv"
	"github.com/tidwall/gjson"
)

type Engine struct {
	trees methodTrees
	lru   *simplelru.LRU
}

func saveResponse(fullPath interface{}, response interface{}) {
	filename := fmt.Sprintf("%x", md5.Sum(bytesconv.StringToBytes(fullPath.(string))))
	fp := filepath.Join(resPath, filename)
	f, err := os.Create(fp)
	if err != nil {
		log.Fatalf("Failed to create response file '%s': %s", fullPath, err)
	}
	defer f.Close()
	ioutil.WriteFile(fp, response.([]byte), 0644)
}

func New() *Engine {
	lru, err := simplelru.NewLRU(128, saveResponse)
	if err != nil {
		log.Fatalf("Failed to create LRU cache: %v", err)
	}
	engine := &Engine{
		trees: make(methodTrees, 0, 9),
		lru:   lru,
	}

	return engine
}

func (engine *Engine) addRoute(method, path string, response []byte) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(response != nil, "response can not be nil")

	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, response)
	engine.lru.Add(path, response)
}

func (engine *Engine) Run(addr ...string) (err error) {
	address := resolveAddress(addr)
	err = http.ListenAndServe(address, engine)
	return
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	httpMethod := req.Method
	rPath := req.URL.Path
	if rPath == "/mock" && httpMethod == "POST" {
		body := req.Body
		defer body.Close()
		code, response := engine.mock(body)
		w.WriteHeader(code)
		w.Write(response)
		return
	}
	code, contentType, response := engine.getValue(httpMethod, rPath)

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	w.Write(response)
}

func (engine *Engine) getValue(method, path string) (code int, contentType string, response []byte) {
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != method {
			continue
		}
		root := t[i].root

		exist := root.getValue(path)
		if exist {
			response = engine.getResponse(path)
			code = 200
			contentType = "application/json"
			return
		}
		break
	}

	return 404, "text/plain", []byte("404 page not found")
}

func (engine *Engine) mock(body io.ReadCloser) (code int, response []byte) {
	bs, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatalf("Failed to read request body: %s", err)
		return 500, []byte("system error")
	}
	method := gjson.GetBytes(bs, "method")
	if !method.Exists() {
		return 500, []byte("method not found")
	}
	path := gjson.GetBytes(bs, "path")
	if !path.Exists() {
		return 500, []byte("path not found")
	}
	resp := gjson.GetBytes(bs, "response")
	if !resp.Exists() {
		return 500, []byte("response not found")
	}
	engine.addRoute(method.String(), path.String(), bytesconv.StringToBytes(resp.String()))
	return 200, []byte("success")
}

func (engine *Engine) getResponse(path string) []byte {
	if res, ok := engine.lru.Get(path); ok {
		return res.([]byte)
	}

	response, err := loadResponseFromLocalFile(path)
	if err != nil {
		log.Fatalf("Failed to load '%s' response from local file: %s", path, err)
		response = []byte("system error")
		return response
	}
	engine.lru.Add(path, response)
	return response
}

func loadResponseFromLocalFile(path string) (response []byte, err error) {
	filename := fmt.Sprintf("%x", md5.Sum(bytesconv.StringToBytes("/hi")))
	response, err = ioutil.ReadFile("responses/" + filename)
	return
}
