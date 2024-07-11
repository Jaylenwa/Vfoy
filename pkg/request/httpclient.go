package request

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"sync"
	"time"
)

type HttpClientInterface interface {
	Get(ctx context.Context, url string, headers map[string]string) (response *http.Response, err error)
	Put(ctx context.Context, url string, headers map[string]string, reqBody []byte) (response *http.Response, err error)
	Post(ctx context.Context, url string, headers map[string]string, body io.Reader) (response *http.Response, err error)
}
type HttpClient struct {
	client *http.Client
}

var (
	httpClientOnce sync.Once
	HttpClientImpl HttpClientInterface
)

func NewHttpClient() HttpClientInterface {
	httpClientOnce.Do(func() {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			MaxIdleConns:    100,
			IdleConnTimeout: time.Minute,
		}
		HttpClientImpl = &HttpClient{
			client: &http.Client{Transport: tr}}
	})
	return HttpClientImpl
}

var _ HttpClientInterface = &HttpClient{}

func (httpClient *HttpClient) Get(ctx context.Context, url string, headers map[string]string) (response *http.Response, err error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	request = request.WithContext(ctx)
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response, err = httpClient.client.Do(request)

	return
}

func (httpClient *HttpClient) Put(ctx context.Context, url string, headers map[string]string, reqBody []byte) (response *http.Response, err error) {
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}
	request = request.WithContext(ctx)
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response, err = httpClient.client.Do(request)

	return
}
func (httpClient *HttpClient) Post(ctx context.Context, url string, headers map[string]string, body io.Reader) (response *http.Response, err error) {
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return
	}
	request = request.WithContext(ctx)
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response, err = httpClient.client.Do(request)
	if err != nil {
		return
	}
	return
}
