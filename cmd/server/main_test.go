package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryMatching(t *testing.T) {
	data := []struct {
		query    string
		expected Query
	}{
		{
			"https://steamcommunity.com/profiles/76561198115172591",
			Query{SteamID64: "76561198115172591"},
		},
		{
			"[U:1:154906863]",
			Query{SteamID3: "[U:1:154906863]"},
		},
	}

	for _, d := range data {
		q, err := parseSteamQuery(d.query)
		assert.NoError(t, err)
		assert.Equal(t, d.expected, q)
	}
}
