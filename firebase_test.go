package firebase

import (
	"testing"
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

func TestValue(t *testing.T) {
	client := new(F)
	client.Init(testUrl, testAuth, nil)

	r := client.Value()

	if r == nil {
		t.Fatalf("No values returned from the server\n")
	}
}

func TestChild(t *testing.T) {
	client := new(F)
	client.Init(testUrl, testAuth, nil)

	r := client.Child("", nil, nil)

	if r == nil {
		t.Fatalf("No child returned from the server\n")
	}
}

func TestPush(t *testing.T) {
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
	c1 := new(F)
	c1.Init(testUrl+"/users", testAuth, nil)

	name := &Name{First: "First", Last: "last"}
	c2, _ := c1.Push(name, nil)

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
	c1 := new(F)
	c1.Init(testUrl+"/users", testAuth, nil)

	name := &Name{First: "First", Last: "last"}
	c2, _ := c1.Push(name, nil)

	newName := &Name{Last: "NewLast"}
	err := c2.Update("", newName, nil)

	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

func TestRemovet(t *testing.T) {
	c1 := new(F)
	c1.Init(testUrl+"/users", testAuth, nil)

	name := &Name{First: "First", Last: "last"}
	c2, _ := c1.Push(name, nil)

	err := c2.Remove("", nil)

	if err != nil {
		t.Fatalf("%v\n", err)
	}
}
