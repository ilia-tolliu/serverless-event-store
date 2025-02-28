package eshttp

import (
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	baseUrl    url.URL
	httpClient http.Client
}

func NewClient(baseUrl string) *Client {
	esUrl, err := url.Parse(baseUrl)
	if err != nil {
		panic(fmt.Sprint("failed to parse base url: ", baseUrl))
	}

	httpClient := http.Client{}

	return &Client{baseUrl: *esUrl, httpClient: httpClient}
}
