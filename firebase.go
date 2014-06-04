// Package firebase impleements a RESTful client for Firebase.
package firebase

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Api is the interface for interacting with Firebase.
// Consumers of this package can mock this interface for testing purposes.
type Api interface {
	Call(method, path, auth string, body []byte, params map[string]string) ([]byte, error)
}

// Client is the Firebase client.
type Client struct {
	// Url is the client's base URL used for all calls.
	Url string

	// Auth is authentication token used when making calls.
	// The token is optional and can also be overwritten on an individual
	// call basis via params.
	Auth string

	// api is the underlying client used to make calls.
	api Api

	// value is the value of the object at the current Url
	value interface{}
}

// Rules is the structure for security rules.
type Rules map[string]interface{}

// f is the internal implementation of the Firebase API client.
type f struct{}

// suffix is the Firebase suffix for invoking their API via HTTP
const suffix = ".json"

var (
	connectTimeout   = time.Duration(10 * time.Second) // timeout for http connection
	readWriteTimeout = time.Duration(10 * time.Second) // timeout for http read/write
)

// httpClient is the HTTP client used to make calls to Firebase
var httpClient = newTimeoutClient(connectTimeout, readWriteTimeout)

// Init initializes the Firebase client with a given root url and optional auth token.
// The initialization can also pass a mock api for testing purposes.
func (c *Client) Init(root, auth string, api Api) {
	if api == nil {
		api = new(f)
	}

	c.api = api
	c.Url = root
	c.Auth = auth
}

// Value returns the value of of the current Url.
func (c *Client) Value() interface{} {
	// if we have not yet performed a look-up, do it so a value is returned
	if c.value == nil {
		var v interface{}
		c = c.Child("", nil, v)
	}

	if c == nil {
		return nil
	}

	return c.value
}

// Child returns a populated pointer for a given path.
// If the path cannot be found, a null pointer is returned.
func (c *Client) Child(path string, params map[string]string, v interface{}) *Client {
	u := c.Url + "/" + path

	res, err := c.api.Call("GET", u, c.Auth, nil, params)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(res, &v)
	if err != nil {
		log.Printf("%v\n", err)
		return nil
	}

	ret := &Client{
		api:   c.api,
		Auth:  c.Auth,
		Url:   u,
		value: v}

	return ret
}

// Push creates a new value under the current root url.
// A populated pointer with that value is also returned.
func (c *Client) Push(value interface{}, params map[string]string) (*Client, error) {
	body, err := json.Marshal(value)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	res, err := c.api.Call("POST", c.Url, c.Auth, body, params)
	if err != nil {
		return nil, err
	}

	var r map[string]string

	err = json.Unmarshal(res, &r)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	ret := &Client{
		api:   c.api,
		Auth:  c.Auth,
		Url:   c.Url + "/" + r["name"],
		value: value}

	return ret, nil
}

// Set overwrites the value at the specified path and returns populated pointer
// for the updated path.
func (c *Client) Set(path string, value interface{}, params map[string]string) (*Client, error) {
	u := c.Url + "/" + path

	body, err := json.Marshal(value)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	res, err := c.api.Call("PUT", u, c.Auth, body, params)

	if err != nil {
		return nil, err
	}

	ret := &Client{
		api:  c.api,
		Auth: c.Auth,
		Url:  u}

	if len(res) > 0 {
		var r interface{}

		err = json.Unmarshal(res, &r)
		if err != nil {
			log.Printf("%v\n", err)
			return nil, err
		}

		ret.value = r
	}

	return ret, nil
}

// Update performs a partial update with the given value at the specified path.
func (c *Client) Update(path string, value interface{}, params map[string]string) error {
	body, err := json.Marshal(value)
	if err != nil {
		log.Printf("%v\n", err)
		return err
	}

	_, err = c.api.Call("PATCH", c.Url+"/"+path, c.Auth, body, params)

	// if we've just updated the root node, clear the value so it gets looked up
	// again and populated correctly since we just applied a diffgram
	if len(path) == 0 {
		c.value = nil
	}

	return err
}

// Remove deletes the data at the given path.
func (c *Client) Remove(path string, params map[string]string) error {
	_, err := c.api.Call("DELETE", c.Url+"/"+path, c.Auth, nil, params)

	return err
}

// Rules returns the security rules for the database.
func (c *Client) Rules(params map[string]string) (Rules, error) {
	res, err := c.api.Call("GET", c.Url+"/.settings/rules", c.Auth, nil, params)
	if err != nil {
		return nil, err
	}

	var v Rules
	err = json.Unmarshal(res, &v)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	return v, nil
}

// SetRules overwrites the existing security rules with the new rules given.
func (c *Client) SetRules(rules *Rules, params map[string]string) error {
	body, err := json.Marshal(rules)
	if err != nil {
		log.Printf("%v\n", err)
		return err
	}

	_, err = c.api.Call("PUT", c.Url+"/.settings/rules", c.Auth, body, params)

	return err
}

// Call invokes the appropriate HTTP method on a given Firebase URL.
func (f *f) Call(method, path, auth string, body []byte, params map[string]string) ([]byte, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	path += suffix
	qs := url.Values{}

	// if the client has an auth, set it as a query string.
	// the caller can also override this on a per-call basis
	// which will happen via params below
	if len(auth) > 0 {
		qs.Set("auth", auth)
	}

	for k, v := range params {
		qs.Set(k, v)
	}

	if len(qs) > 0 {
		path += "?" + qs.Encode()
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(body))
	if err != nil {
		log.Printf("Cannot create Firebase request: %v\n", err)
		return nil, err
	}

	req.Close = true
	log.Printf("Calling %v %q\n", method, path)

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Request to Firebase failed: %v\n", err)
		return nil, err
	}
	defer res.Body.Close()

	ret, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Cannot parse Firebase response: %v\n", err)
		return nil, err
	}

	if res.StatusCode >= 400 {
		err = errors.New(string(ret))
		log.Printf("Error encountered from Firebase: %v\n", err)
		return nil, err
	}

	return ret, nil
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func newTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}
