package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/plally/steamid.id/internal/db"
	"github.com/plally/steamid.id/internal/routes"
	"github.com/plally/steamid.id/internal/steamapi"
)

func main() {
	db := db.NewRedis()
	defer db.Close()

	slog.Info("starting server")
	steamAPI := steamapi.New(os.Getenv("STEAM_API_KEY"))

	r := routes.GetRouter(steamAPI, &db)

	os.WriteFile("buildtime.txt", []byte(time.Now().Format(time.RFC1123)), 0644)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
