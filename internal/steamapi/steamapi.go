package steamapi

import (
	"encoding/json"
	"fmt"
	"time"

	"net/http"
)

type SteamAPI struct {
	APIKey string
	client http.Client
}

func New(apiKey string) *SteamAPI {
	return &SteamAPI{
		APIKey: apiKey,

		client: http.Client{
			Timeout: time.Second * 10,
		},
	}
}

type VanityURLResponse struct {
	Response struct {
		SteamID string `json:"steamid"`
		Success int    `json:"success"`
	} `json:"response"`
}

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

// ResolveVanityURL description
func (s *SteamAPI) ResolveVanityURL(vanityURL string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=%s&vanityurl=%s", s.APIKey, vanityURL), nil)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("steam api returned %d", resp.StatusCode)
	}

	var vanityURLResponse VanityURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&vanityURLResponse); err != nil {
		return "", err
	}

	if vanityURLResponse.Response.Success != 1 {
		return "", fmt.Errorf("steam api returned %d", vanityURLResponse.Response.Success)
	}

	return vanityURLResponse.Response.SteamID, nil
}

func (s *SteamAPI) GetPlayerSummary(steamID string) (Player, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", s.APIKey, steamID), nil)
	if err != nil {
		return Player{}, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return Player{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Player{}, fmt.Errorf("steam api returned %d", resp.StatusCode)
	}

	var getPlayerSummariesResponse GetPlayerSummariesResponse
	if err := json.NewDecoder(resp.Body).Decode(&getPlayerSummariesResponse); err != nil {
		return Player{}, err
	}

	if len(getPlayerSummariesResponse.Response.Players) == 0 {
		return Player{}, fmt.Errorf("steam api returned no players")
	}

	return getPlayerSummariesResponse.Response.Players[0], nil
}
