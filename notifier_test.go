package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

type handlerDescriber func(w http.ResponseWriter, r *http.Request)

func createMockWebserver(uri string, handler handlerDescriber) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("/%s", uri), handler)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	// Sleep here is required because the server needs a moment to boot up, it can cause a race condition, 1 second should be enough between the client and the server.
	time.Sleep(time.Second)

	return srv

}

func TestGet(t *testing.T) {

	var handler handlerDescriber = func(w http.ResponseWriter, r *http.Request) {

		res := Response{
			CurrentVersion:     "1.0.0",
			CurrentDownloadURL: "foo.com",
			Outdated:           true,
			Notifications: []*Notification{
				{12345, "message"},
				{12345, "message"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(res)

	}

	server := createMockWebserver("api/v1/latest-version/test", handler)

	params := &UpdaterParams{
		Application: "test",
	}
	requestSetting := RequestSetting{
		Host: "http://127.0.0.1:5000",
	}
	response, err := Get(params, requestSetting)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response.CurrentVersion != "1.0.0" {
		t.Fatalf("unexpected latest version value: got %s want %s", response.CurrentVersion, "1.0.0")
	}

	if response.CurrentDownloadURL != "foo.com" {
		t.Fatalf("unexpected latest release date value: got %s want %s", response.CurrentDownloadURL, "foo.com")
	}

	if response.Outdated != true {
		t.Fatalf("unexpected outdated value: got %t want %t", response.Outdated, true)
	}

	if len(response.Notifications) != 2 {
		t.Fatalf("unexpected version count: got %d want %d", len(response.Notifications), 2)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.Shutdown(ctx)

}
func TestRequestTimeputGet(t *testing.T) {

	var handler handlerDescriber = func(w http.ResponseWriter, r *http.Request) {

		time.Sleep(4 * time.Second)
		res := Response{}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(res)

	}

	server := createMockWebserver("api/v1/latest-version/test", handler)

	params := &UpdaterParams{
		Application: "test",
	}
	requestSetting := RequestSetting{
		Host: "http://127.0.0.1:5000",
	}

	_, err := Get(params, requestSetting)

	if err == nil {
		t.Errorf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.Shutdown(ctx)

}

func TestCheckInterval(t *testing.T) {

	callerCount := 0
	ctx, cancelFn := context.WithCancel(context.Background())
	params := &UpdaterParams{
		Application: "test",
	}
	requestSetting := RequestSetting{
		Host: "http://127.0.0.1:5000/test",
	}

	var update = func(*Response, error) {
		callerCount++
	}

	interval := 1 * time.Second
	GetInterval(ctx, params, interval, update, requestSetting)

	time.Sleep(time.Second * 5)
	cancelFn()

	if callerCount != 4 {
		t.Fatalf("unexpected caller function: got %d want %d", callerCount, 4)
	}
}
