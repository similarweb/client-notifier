package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// HTTPDefaultHost default of webserver
	HTTPDefaultHost = "https://version.similarweb.engineering"

	// HTTRequestTimeout defins the http timeout
	HTTRequestTimeout = 3
)

// RequestSetting request setting
type RequestSetting struct {
	Host string
}

// UpdaterParams are get parameters for notifier HTTP request.
type UpdaterParams struct {
	// Name the application
	Application string

	// Name of the component
	Component string

	// Application/component versions
	Version string
}

// Response is the response from notifier webserver.
type Response struct {
	CurrentVersion     string          `json:"current_version"`
	CurrentDownloadURL string          `json:"current_download_url"`
	Outdated           bool            `json:"outdated"`
	Notifications      []*Notification `json:"notifications"`
}

// Notification is a Notification message from notifier webserver.
type Notification struct {
	Date    int    `json:"date"`
	Message string `json:"message"`
}

// Get creates http call fo getting the latest version of the application
func Get(p *UpdaterParams, requestSetting RequestSetting) (*Response, error) {

	client := &http.Client{
		Timeout: HTTRequestTimeout * time.Second,
	}

	data := url.Values{}
	data.Set("component", p.Component)

	host := HTTPDefaultHost

	if requestSetting.Host != "" {
		host = requestSetting.Host
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/latest-version/%s", host, p.Application), bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r io.Reader = resp.Body

	var result Response
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil

}

// GetInterval called to get the the version of the application during the inteval time.
func GetInterval(ctx context.Context, p *UpdaterParams, interval time.Duration, update func(*Response, error), requestSetting RequestSetting) {

	go func() {
		for {
			select {
			case <-time.After(interval):
				resp, err := Get(p, requestSetting)
				update(resp, err)
			case <-ctx.Done():
				return
			}
		}
	}()
}
