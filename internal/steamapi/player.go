package steamapi

import (
	"context"
	"strings"
	"time"

	"log/slog"
)

type Player struct {
	Avatar                   string  `json:"avatar,omitempty"`
	Avatarfull               string  `json:"avatarfull,omitempty"`
	Avatarhash               string  `json:"avatarhash,omitempty"`
	Avatarmedium             string  `json:"avatarmedium,omitempty"`
	Communityvisibilitystate float64 `json:"communityvisibilitystate,omitempty"`
	Loccityid                float64 `json:"loccityid,omitempty"`
	Loccountrycode           string  `json:"loccountrycode,omitempty"`
	Locstatecode             string  `json:"locstatecode,omitempty"`
	PersonaName              string  `json:"personaname,omitempty"`
	Personastate             float64 `json:"personastate,omitempty"`
	Personastateflags        float64 `json:"personastateflags,omitempty"`
	Primaryclanid            string  `json:"primaryclanid,omitempty"`
	Profilestate             float64 `json:"profilestate,omitempty"`
	ProfileURL               string  `json:"profileurl,omitempty"`
	Realname                 string  `json:"realname,omitempty"`
	Steamid                  string  `json:"steamid,omitempty"`
	Timecreated              float64 `json:"timecreated,omitempty"`
}

type GetPlayerSummariesResponse struct {
	Response struct {
		Players []Player `json:"players,omitempty"`
	} `json:"response,omitempty"`
}

type summaryRequest struct {
	steamID string
	result  chan<- Player
}

const MAX_SUMMARY_REQUESTS = 100

func (s *SteamAPI) startNewWorker() {
	go s.steamSummaryJob(s.steamSummaryRequests)
}

func (s *SteamAPI) steamSummaryJob(requets <-chan summaryRequest) {
	requests := make([]summaryRequest, 0, MAX_SUMMARY_REQUESTS)

	for {
		select {
		case <-s.done:
			return
		case <-time.After(time.Millisecond * 100):
			if len(requests) > 0 {
				go s.doBatchRequest(requests)
				requests = make([]summaryRequest, 0, MAX_SUMMARY_REQUESTS)
			}
		case req := <-requets:
			requests = append(requests, req)
			if len(requests) >= cap(requests) {
				go s.doBatchRequest(requests)
				requests = make([]summaryRequest, 0, MAX_SUMMARY_REQUESTS)
			}
		}

	}
}

func (s *SteamAPI) doBatchRequest(requests []summaryRequest) {
	steamIDs := make([]string, 0, len(requests))
	for _, req := range requests {
		steamIDs = append(steamIDs, req.steamID)
	}

	plys, err := s.GetPlayerSummaries(steamIDs)
	if err != nil {
		slog.With("err", err).Error("failed to get player summaries")
	}

	for _, req := range requests {
		req.result <- plys[req.steamID]
	}
}

func (s *SteamAPI) GetPlayerSummary(ctx context.Context, steamID string) (Player, error) {
	c := make(chan Player)
	req := summaryRequest{
		steamID: steamID,
		result:  c,
	}

	select {
	case <-ctx.Done():
		return Player{}, ctx.Err()
	case s.steamSummaryRequests <- req:
	}

	select {
	case <-ctx.Done():
		return Player{}, ctx.Err()
	case ply := <-c:
		return ply, nil
	}
}

func (s *SteamAPI) GetPlayerSummaries(steamIDs []string) (map[string]Player, error) {
	playersResponse := make(map[string]Player)

	steamIDsJoined := strings.Join(steamIDs, ",")
	var getPlayerSummariesResponse GetPlayerSummariesResponse

	err := s.Get("ISteamUser/GetPlayerSummaries/v0002/", map[string]string{
		"steamids": steamIDsJoined,
	}, &getPlayerSummariesResponse)
	if err != nil {
		return nil, err
	}

	for _, ply := range getPlayerSummariesResponse.Response.Players {
		playersResponse[ply.Steamid] = ply
	}

	return playersResponse, nil
}
