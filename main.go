package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"log/slog"

	"github.com/plally/steamid.id/internal/steamapi"
)

func main() {
	slog.Info("starting server")
	tpl, err := template.ParseFS(resources, "public/index.html", "public/user.html", "public/components.html")
	if err != nil {
		log.Fatal(err)
	}
	steamAPI := steamapi.New(os.Getenv("STEAM_API_KEY"))

	r := GetRouter(steamAPI, tpl)

	os.WriteFile("buildtime.txt", []byte(time.Now().Format(time.RFC1123)), 0644)

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}

// regex patterns
var (
	customNamePattern   = regexp.MustCompile(`^([a-zA-Z0-9_-]+)$`)
	customURLPattern    = regexp.MustCompile(`^https?:\/\/steamcommunity\.com\/id\/([a-zA-Z0-9_-]+)\/?$`)
	steamProfilePattern = regexp.MustCompile(`^https?:\/\/steamcommunity\.com\/profiles\/([0-9]+)\/?$`)
	steamID64Pattern    = regexp.MustCompile(`^([0-9]{17})$`)
	steamID32Pattern    = regexp.MustCompile(`^STEAM_([0-1]):([0-1]):([0-9]+)$`)
)

type Query struct {
	CustomURLName string
	SteamID64     string
	SteamID32     string
}

var ErrInvalidQuery = errors.New("invalid query")

func parseSteamQuery(query string) (Query, error) {
	query = strings.TrimSpace(query)
	if customURLPattern.MatchString(query) {
		name := customURLPattern.FindStringSubmatch(query)[1]
		return Query{CustomURLName: name}, nil
	}
	if steamProfilePattern.MatchString(query) {
		id := steamProfilePattern.FindStringSubmatch(query)[1]
		return Query{SteamID64: id}, nil
	}
	if steamID64Pattern.MatchString(query) {
		id := steamID64Pattern.FindStringSubmatch(query)[1]
		return Query{SteamID64: id}, nil
	}
	if steamID32Pattern.MatchString(query) {
		id := steamID32Pattern.FindStringSubmatch(query)[1]
		return Query{SteamID32: id}, nil
	}

	if customNamePattern.MatchString(query) {
		name := customNamePattern.FindStringSubmatch(query)[1]
		return Query{CustomURLName: name}, nil
	}
	return Query{}, ErrInvalidQuery
}
