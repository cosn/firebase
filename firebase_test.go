package firebase

import (
	"testing"
	"time" // this shouldn't be needed, but without a small delay between calls, the Go HttpClient panics.
)

type Name struct {
	First string `json:",omitempty"`
	Last  string `json:",omitempty"`
}

/*
Set the two variables below and set them to your own
Firebase URL and credentials (optional) if you're forking the code
and want to test your changes.
*/

var testUrl, testAuth string

// TODO: report the issue to #GoLang and remove after clarified
const bugDelay = 50

func TestValue(t *testing.T) {
	keysInit()

	client := new(F)
	client.Init(testUrl, testAuth, nil)

	r := client.Value()

	if r == nil {
		t.Fatalf("No values returned from the server\n")
	}
}

func TestChild(t *testing.T) {
	keysInit()
	time.Sleep(bugDelay * time.Millisecond)
	client := new(F)
	client.Init(testUrl, testAuth, nil)

	r := client.Child("", nil, nil)

	if r == nil {
		t.Fatalf("No child returned from the server\n")
	}
}

func TestPush(t *testing.T) {
	keysInit()
	time.Sleep(bugDelay * time.Millisecond)
	client := new(F)
	client.Init(testUrl, testAuth, nil)

	name := &Name{First: "FirstName", Last: "LastName"}

	r, err := client.Push(name, nil)

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if r == nil {
		t.Fatalf("No client returned from the server\n")
	}
}

func TestSet(t *testing.T) {
	keysInit()
	time.Sleep(bugDelay * time.Millisecond)
	c1 := new(F)
	c1.Init(testUrl+"/users", testAuth, nil)

	name := &Name{First: "First", Last: "last"}
	c2, _ := c1.Push(name, nil)

	time.Sleep(bugDelay * time.Millisecond)
	newName := &Name{First: "NewFirst", Last: "NewLast"}
	r, err := c2.Set("", newName, map[string]string{"print": "silent"})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if r == nil {
		t.Fatalf("No client returned from the server\n")
	}
}

func TestUpdate(t *testing.T) {
	keysInit()
	time.Sleep(bugDelay * time.Millisecond)
	c1 := new(F)
	c1.Init(testUrl+"/users", testAuth, nil)

	name := &Name{First: "First", Last: "last"}
	c2, _ := c1.Push(name, nil)

	time.Sleep(bugDelay * time.Millisecond)
	newName := &Name{Last: "NewLast"}
	err := c2.Update("", newName, nil)

	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestRemovet(t *testing.T) {
	keysInit()
	time.Sleep(bugDelay * time.Millisecond)
	c1 := new(F)
	c1.Init(testUrl+"/users", testAuth, nil)

	name := &Name{First: "First", Last: "last"}
	c2, _ := c1.Push(name, nil)

	time.Sleep(bugDelay * time.Millisecond)
	err := c2.Remove("", nil)

	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func keysInit() {
	if len(testUrl) == 0 {
		testUrl = keyUrl
		testAuth = keyAuth
	}
}
