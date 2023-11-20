package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testQueryMatching(t *testing.T) {
	data := []struct {
		query    string
		expected Query
	}{
		{
			"https://steamcommunity.com/profiles/76561198115172591",
			Query{SteamID64: "76561198115172591"},
		},
	}

	for _, d := range data {
		q, err := parseSteamQuery(d.query)
		assert.NoError(t, err)
		assert.Equal(t, d.expected, q)
	}
}
