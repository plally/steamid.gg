package matching

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/plally/steamid"
	"github.com/plally/steamid.id/internal/steamapi"
)

var (
	customNamePattern   = regexp.MustCompile(`^([a-zA-Z0-9_-]+)$`)
	customURLPattern    = regexp.MustCompile(`^https?:\/\/steamcommunity\.com\/id\/([a-zA-Z0-9_-]+)\/?$`)
	steamProfilePattern = regexp.MustCompile(`^https?:\/\/steamcommunity\.com\/profiles\/([0-9]+)\/?$`)
	steamID64Pattern    = regexp.MustCompile(`^([0-9]{17})$`)
	steamID32Pattern    = regexp.MustCompile(`^STEAM_([0-1]):([0-1]):([0-9]+)$`)
	stemID3Pattern      = regexp.MustCompile(`^\[U:1:[0-9]+\]$`)
)

type Query struct {
	CustomURLName string
	SteamID64     string
	SteamID32     string
	SteamID3      string
}

var ErrInvalidQuery = errors.New("invalid query")

func customNameResolver(api *steamapi.SteamAPI, query string) (steamid.SteamID, error) {
	steamID, err := api.ResolveVanityURL(query)
	if err != nil {
		return 0, err
	}

	if steamID.Success != steamapi.VanityURLSuccess {
		return 0, nil
	}

	return steamid.SteamID64(steamID.SteamID)
}

func discardAPI(f func(query string) (steamid.SteamID, error)) func(api *steamapi.SteamAPI, query string) (steamid.SteamID, error) {
	return func(_ *steamapi.SteamAPI, query string) (steamid.SteamID, error) {
		return f(query)
	}
}

var possibleQueries = []struct {
	name    string
	pattern *regexp.Regexp
	index   int
	f       func(api *steamapi.SteamAPI, query string) (steamid.SteamID, error)
}{
	{
		name:    "Custom URL",
		pattern: regexp.MustCompile(`^https?:\/\/steamcommunity\.com\/id\/([a-zA-Z0-9_-]+)\/?$`),
		index:   1,
		f:       customNameResolver,
	},
	{
		name:    "Steam Profile URL",
		pattern: regexp.MustCompile(`^https?:\/\/steamcommunity\.com\/profiles\/([0-9]+)\/?$`),
		index:   1,
		f:       discardAPI(steamid.SteamID64),
	},
	{
		name:    "SteamID64",
		pattern: regexp.MustCompile(`^[0-9]{17}$`),
		index:   0,
		f:       discardAPI(steamid.SteamID64),
	},
	{
		name:    "SteamID32",
		pattern: regexp.MustCompile(`^STEAM_[0-1]:[0-1]:[0-9]+$`),
		index:   0,
		f:       discardAPI(steamid.SteamID32),
	},
	{
		name:    "SteamID3",
		pattern: regexp.MustCompile(`^\[U:1:[0-9]+\]$`),
		index:   0,
		f:       discardAPI(steamid.SteamID3),
	},
	{
		name:    "Custom URL Name",
		pattern: regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		index:   0,
		f:       customNameResolver,
	},
}

type ParseRequest struct {
	Query string
	API   *steamapi.SteamAPI
}

type ParseResponse struct {
	Name    string
	SteamID steamid.SteamID
}

func ParseSteamQuery(ctx context.Context, req ParseRequest) (ParseResponse, error) {
	query := strings.TrimSpace(req.Query)

	for _, pq := range possibleQueries {
		matches := pq.pattern.FindStringSubmatch(query)
		if matches != nil && len(matches) >= pq.index {
			steamid, err := pq.f(req.API, matches[pq.index])
			return ParseResponse{
				Name:    pq.name,
				SteamID: steamid,
			}, err

		}
	}
	return ParseResponse{}, nil
}
