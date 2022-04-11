package gmock

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/takiya562/go-mock/internal/bytesconv"
)

func fakeBody() io.ReadCloser {
	body := `{"method": "GET", "path": "/hi", "response": {"status": 200, "body": "hi"}}`
	return io.NopCloser(strings.NewReader(body))
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

func TestMock(t *testing.T) {
	engine := New()
	code, _ := engine.mock(fakeBody())
	if code != 200 {
		t.Errorf("Failed to mock: %d", code)
	}
	code, _, response := engine.getValue("GET", "/hi")
	if code != 200 {
		t.Errorf("Failed to get value: %d", code)
	}
	if string(response) != `{"status": 200, "body": "hi"}` {
		t.Errorf("Response mismatch: %s", string(response))
	}
}

func TestMockMissMethod(t *testing.T) {
	engine := New()
	body := io.NopCloser(strings.NewReader(`{"path": "/hi", "response": {"status": 200, "body": "hi"}}`))
	code, _ := engine.mock(body)
	if code != 500 {
		t.Errorf("mock miss method code mismatch: %d", code)
	}
}

func TestMockMissPath(t *testing.T) {
	engine := New()
	body := io.NopCloser(strings.NewReader(`{"method": "GET", "response": {"status": 200, "body": "hi"}}`))
	code, _ := engine.mock(body)
	if code != 500 {
		t.Errorf("mock miss path code mismatch: %d", code)
	}
}

func TestMockMissResponse(t *testing.T) {
	engine := New()
	body := io.NopCloser(strings.NewReader(`{"method": "GET", "path": "\hi"}`))
	code, _ := engine.mock(body)
	if code != 500 {
		t.Errorf("mock miss response code mismatch: %d", code)
	}
}

func TestGetResponseFromFile(t *testing.T) {
	engine := New()
	bs := engine.getResponse("/hi")
	if string(bs) != "hi" {
		t.Errorf("Response mismatch: %s", string(bs))
	}
}
