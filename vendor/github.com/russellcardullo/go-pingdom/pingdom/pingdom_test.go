package pingdom

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mux    *http.ServeMux
	client *Client
	server *httptest.Server
)

func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// test client
	client, _ = NewClientWithConfig(ClientConfig{
		APIToken: "my_api_key",
	})

	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	assert.Equal(t, want, r.Method)
}

func TestNewClientWithConfig(t *testing.T) {
	c, err := NewClientWithConfig(ClientConfig{
		APIToken: "key",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
}

func TestNewClientWithEnvAPITokenDoesNotOverride(t *testing.T) {
	os.Setenv("PINGDOM_API_TOKEN", "envSetAwesome")
	defer os.Unsetenv("PINGDOM_API_TOKEN")
	c, err := NewClientWithConfig(ClientConfig{
		APIToken: "key",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, c.APIToken, "key")
}

func TestNewClientWithEnvAPITokenWorks(t *testing.T) {
	os.Setenv("PINGDOM_API_TOKEN", "envSetAwesome")
	defer os.Unsetenv("PINGDOM_API_TOKEN")
	c, err := NewClientWithConfig(ClientConfig{})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, c.APIToken, "envSetAwesome")
}

func TestNewRequest(t *testing.T) {
	setup()
	defer teardown()

	req, err := client.NewRequest("GET", "/checks", nil)

	assert.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, client.BaseURL.String()+"/checks", req.URL.String())
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	body := new(foo)
	want := &foo{"a"}

	_, err := client.Do(req, body)
	assert.NoError(t, err)
	assert.Equal(t, want, body)
}

func TestValidateResponse(t *testing.T) {
	valid := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader("OK")),
	}

	assert.NoError(t, validateResponse(valid))

	invalid := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body: ioutil.NopCloser(strings.NewReader(`{
			"error" : {
				"statuscode": 400,
				"statusdesc": "Bad Request",
				"errormessage": "This is an error"
			}
		}`)),
	}

	want := &PingdomError{400, "Bad Request", "This is an error"}
	assert.Equal(t, want, validateResponse(invalid))
}
