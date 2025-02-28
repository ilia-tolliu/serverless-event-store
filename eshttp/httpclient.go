package eshttp

import (
	"fmt"
	"net/http"
	"net/url"
)

type EsHttpClient struct {
	baseUrl    url.URL
	httpClient http.Client
}

func New(baseUrl string) *EsHttpClient {
	esUrl, err := url.Parse(baseUrl)
	if err != nil {
		panic(fmt.Sprint("failed to parse base url: ", baseUrl))
	}

	httpClient := http.Client{}

	return &EsHttpClient{baseUrl: *esUrl, httpClient: httpClient}
}
