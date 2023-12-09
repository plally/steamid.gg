package steamapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"net/http"
	"net/url"
)

type SteamAPI struct {
	APIKey  string
	client  http.Client
	BaseURL string

	done                 chan struct{}
	steamSummaryRequests chan summaryRequest
}

func New(apiKey string) *SteamAPI {
	api := &SteamAPI{
		APIKey:  apiKey,
		BaseURL: "http://api.steampowered.com/",
		client: http.Client{
			Timeout: time.Second * 10,
		},

		done:                 make(chan struct{}),
		steamSummaryRequests: make(chan summaryRequest, MAX_SUMMARY_REQUESTS*2),
	}

	go api.startNewWorker()
	return api
}

func (s *SteamAPI) Close() {
	close(s.done)
}

func (s *SteamAPI) Get(path string, params map[string]string, out any) error {
	reqUrl, err := url.Parse(fmt.Sprintf("%s/%s", s.BaseURL, path))
	if err != nil {
		return err
	}

	q := reqUrl.Query()
	q.Set("key", s.APIKey)
	for k, v := range params {
		q.Set(k, v)
	}
	reqUrl.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", reqUrl.String(), nil)
	if err != nil {
		return err
	}

	slog.With("url", reqUrl.String()).Info("Doing GET request")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("steam api returned %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}

	return nil
}

type Response[T any] struct {
	Response T `json:"response"`
}
