package main

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"log/slog"

	"github.com/plally/steamid.id/internal/steamapi"
)

//go:embed public/*
var resources embed.FS

func main() {
	slog.Info("starting server")
	tpl, err := template.ParseFS(resources, "public/index.html", "public/user.html")
	if err != nil {
		log.Fatal(err)
	}
	steamAPI := steamapi.New(os.Getenv("STEAM_API_KEY"))

	r := GetRouter(steamAPI, tpl)

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

// copied from https://github.com/MrWaggel/gosteamconv/tree/master
func steamID32to64(steamString string) (string, error) {
	Y, err := strconv.Atoi(steamString[8:9])
	if err != nil {
		return "", err
	}

	Z, err := strconv.Atoi(steamString[10:])
	if err != nil {
		return "", err
	}
	i := int64((Z * 2) + 76561197960265728 + Y)

	return strconv.FormatInt(i, 10), nil
}
func steamID64to32(steam64 string) (string, error) {
	steamInt, err := strconv.ParseInt(steam64, 10, 64)
	if err != nil {
		return "", errors.New("failed to parse steamid64")
	}

	if steamInt <= 76561197960265728 {
		return "", errors.New("steamid too small")
	}

	steamInt = steamInt - 76561197960265728
	remainder := steamInt % 2
	steamInt = steamInt / 2
	return "STEAM_0:" + strconv.FormatInt(remainder, 10) + ":" + strconv.FormatInt(steamInt, 10), nil

}

func steamID64toSteamID3(steam64 string) (string, error) {
	steamInt, err := strconv.ParseInt(steam64, 10, 64)
	if err != nil {
		return "", errors.New("failed to parse steamid64")
	}

	if steamInt <= 76561197960265728 {
		return "", errors.New("steamid too small")
	}

	steamInt = steamInt - 76561197960265728
	return "[U:1:" + strconv.FormatInt(steamInt, 10) + "]", nil
}
