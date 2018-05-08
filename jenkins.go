package jenkins

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	defaultHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

// BuildStatus ...
type BuildStatus int

// BuildState ...
type BuildState int

const (
	// BuildFail ...
	BuildFail BuildStatus = iota
	// BuildSuccess ...
	BuildSuccess
	// BuildUnknown ...
	BuildUnknown
)

const (
	// BuildRunning ...
	BuildRunning BuildState = iota
	// BuildComplete ...
	BuildComplete
)

// Client ...
type Client struct {
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
}

type queueItemResponse struct {
	Executable *executable `json:"executable"`
}

type executable struct {
	URL string `json:"url"`
}

type buildResponse struct {
	Building *bool   `json:"building"`
	Result   *string `json:"result"`
}

// NewClient ...
func NewClient(baseURL, username, token string, httpClient *http.Client) *Client {
	h := httpClient
	if httpClient == nil {
		h = defaultHTTPClient
	}
	return &Client{baseURL: baseURL, username: username, token: token, httpClient: h}

}

// BuildJob ...
func (c *Client) BuildJob(jobPath string) string {
	apiJobPath := fmt.Sprintf("%s/job/%s/build",
		c.baseURL,
		strings.Join(strings.Split(jobPath, "/"), "/job/"),
	)
	resp, err := c.request("POST", apiJobPath, nil)
	logError(err)

	itemURL := &resp.Header["Location"][0]

	return c.getBuildURL(*itemURL)
}

func (c *Client) getBuildURL(itemURL string) string {
	url := fmt.Sprintf("%s/api/json", itemURL)
	resp, err := c.request("GET", url, nil)
	logError(err)

	body, err := ioutil.ReadAll(resp.Body)
	logError(err)
	defer resp.Body.Close()

	queueItemResponse := &queueItemResponse{}
	err = json.Unmarshal([]byte(body), queueItemResponse)
	logError(err)

	buidURL := queueItemResponse.Executable.URL

	return buidURL
}

// BuildStatus ...
func (c *Client) BuildStatus(buildURL string) (BuildState, BuildStatus) {
	url := fmt.Sprintf("%s/api/json", buildURL)
	resp, err := c.request("GET", url, nil)
	logError(err)

	body, err := ioutil.ReadAll(resp.Body)
	logError(err)
	defer resp.Body.Close()

	buildResponse := &buildResponse{}
	err = json.Unmarshal([]byte(body), buildResponse)

	logError(err)

	if *buildResponse.Building {
		return BuildRunning, BuildUnknown
	}

	return BuildComplete, BuildSuccess
}

func (c *Client) request(method, URL string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.token)
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
