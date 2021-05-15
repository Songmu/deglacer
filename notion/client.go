package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	apiBaseUrl    = "https://api.notion.com/"
	notionVersion = "2021-05-13"
)

type Client struct {
	AuthToken  string
	HTTPClient *http.Client
}

func (c Client) RetrievePage(pageId string) (*Page, error) {
	path := "v1/pages/" + pageId
	var page Page
	err := doNotionApi(c, path, "GET", nil, &page)
	if err != nil {
		return nil, err
	}
	fmt.Println("unko!!!!!!!!!!")

	return &page, nil
}

func doNotionApi(c Client, path string, method string, requestData interface{}, result interface{}) error {
	uri := apiBaseUrl + path
	var js []byte
	var err error
	if requestData != nil {
		js, err = json.Marshal(requestData)
		if err != nil {
			return err
		}
	}
	body := bytes.NewBuffer(js)

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return err
	}

	req.Header.Set("Notion-Version", notionVersion)
	if c.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.AuthToken))
	}
	var rsp *http.Response
	httpClient := c.getHTTPClient()
	rsp, err = httpClient.Do(req)
	if err != nil {
		return err
	}

	var d []byte
	d, _ = ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != 200 {
		js, _ = json.Marshal(d)
		return fmt.Errorf("Error: status code %s\nBody:\n%s\n", rsp.Status, js)
	}
	err = json.Unmarshal(d, result)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getHTTPClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	httpClient := *http.DefaultClient
	httpClient.Timeout = time.Second * 30
	return &httpClient
}
