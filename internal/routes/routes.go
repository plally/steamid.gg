package routes

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/plally/steamid"
	"github.com/plally/steamid.id/internal/db"
	"github.com/plally/steamid.id/internal/matching"
	"github.com/plally/steamid.id/internal/responses"
	"github.com/plally/steamid.id/internal/steamapi"
)

//go:embed public/*
var resources embed.FS

type routeState struct {
	steamAPI *steamapi.SteamAPI
	tpl      *template.Template
	db       *db.RedisStore
}

func redirectError(w http.ResponseWriter, r *http.Request, err string) {
	err = url.QueryEscape(err)
	http.Redirect(w, r, fmt.Sprintf("/?error=%s", err), http.StatusSeeOther)
}

func (s *routeState) handlePostSearch(w http.ResponseWriter, r *http.Request) {
	queryString := r.FormValue("search")
	log := slog.With("query", queryString)
	resp, err := matching.ParseSteamQuery(r.Context(), matching.ParseRequest{
		API:   s.steamAPI,
		Query: queryString,
	})
	if err != nil {
		log.With("err", err).Error("failed to parse query")
		redirectError(w, r, "Failed to parse query")
	}

	if resp.SteamID == 0 {
		if resp.Name != "" {
			redirectError(w, r, fmt.Sprintf("Failed to resolve query as %v", resp.Name))
		} else {
			redirectError(w, r, "Failed to parse query, unknown query type")
		}
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/user/%s", resp.SteamID.SteamID64String()), http.StatusSeeOther)
}

type IndexData struct {
	Error  string
	Search string
}

func (s *routeState) handleGetIndex(w http.ResponseWriter, r *http.Request) {
	responses.CacheControlCache6Month(w, r)

	err := s.tpl.ExecuteTemplate(w, "index.html", IndexData{
		Error:  r.URL.Query().Get("error"),
		Search: r.URL.Query().Get("search"),
	})
	if err != nil {
		slog.With("err", err).Error("failed to execute template")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

type PlayerData struct {
	Search      string `json:",omitempty"`
	Username    string
	Avatar      string
	CustomURL   string
	ProfileURL  string
	RealName    string
	SteamID32   string
	SteamID64   string
	SteamID3    string
	Location    string
	CreatedAt   string
	LastUpdated string
	Error       string
}

func (s *routeState) getPlayerSummary(ctx context.Context, steamID64 string) (*db.PlayerData, error) {
	dbData, err := s.db.GetPlayerData(ctx, steamID64)
	if err != nil {
		return nil, err
	}
	if dbData != nil && dbData.Username == "" && dbData.LastUpdated < time.Now().Add(-1*time.Minute).Unix() {
		slog.With("steamid", steamID64).Info("DB data is invalid and cached more than 10 minutes ago, fetching from API")
		dbData = nil
	} else if dbData != nil {
		return dbData, nil
	}

	ply, err := s.steamAPI.GetPlayerSummary(ctx, steamID64)
	if err != nil {
		return nil, err
	}

	data := db.PlayerData{
		Username:  ply.PersonaName,
		Avatar:    ply.Avatarfull,
		CustomURL: ply.ProfileURL,
		RealName:  ply.Realname,
		SteamID64: steamID64,
		Location:  ply.Loccountrycode,
		CreatedAt: int64(ply.Timecreated),

		LastUpdated: time.Now().Unix(),
	}

	err = s.db.SetPlayerData(ctx, steamID64, data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *routeState) getUserAPI(w http.ResponseWriter, r *http.Request) {
	steamID64 := chi.URLParam(r, "steamid")
	log := slog.With("steamid", steamID64)
	ctx := r.Context()
	steamID, err := steamid.SteamID64(steamID64)
	if err != nil {
		log.With("err", err).Error("failed to parse steamid64")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ply, err := s.getPlayerSummary(ctx, steamID64)
	if err != nil {
		log.With("err", err).Error("failed to get player summary")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := PlayerData{
		Username:    ply.Username,
		Avatar:      ply.Avatar,
		CustomURL:   ply.CustomURL,
		ProfileURL:  fmt.Sprintf("https://steamcommunity.com/profiles/%s", ply.SteamID64),
		CreatedAt:   time.Unix(ply.CreatedAt, 0).Format(time.ANSIC),
		RealName:    ply.RealName,
		SteamID32:   steamID.SteamID32String(),
		SteamID64:   steamID.SteamID64String(),
		Location:    ply.Location,
		SteamID3:    steamID.SteamID3String(),
		LastUpdated: time.Unix(ply.LastUpdated, 0).Format(time.RFC3339),
	}

	responses.CacheControlCacheForWeek(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *routeState) getUser(w http.ResponseWriter, r *http.Request) {
	steamID64 := chi.URLParam(r, "steamid")
	log := slog.With("steamid", steamID64)
	ctx := r.Context()
	steamID, err := steamid.SteamID64(steamID64)
	if err != nil {
		log.With("err", err).Error("failed to parse steamid64")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ply, err := s.getPlayerSummary(ctx, steamID64)
	if err != nil {
		log.With("err", err).Error("failed to get player summary")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := PlayerData{
		Username:    ply.Username,
		Avatar:      ply.Avatar,
		CustomURL:   ply.CustomURL,
		ProfileURL:  fmt.Sprintf("https://steamcommunity.com/profiles/%s", ply.SteamID64),
		CreatedAt:   time.Unix(ply.CreatedAt, 0).Format(time.ANSIC),
		RealName:    ply.RealName,
		SteamID32:   steamID.SteamID32String(),
		SteamID64:   steamID.SteamID64String(),
		Location:    ply.Location,
		SteamID3:    steamID.SteamID3String(),
		LastUpdated: time.Unix(ply.LastUpdated, 0).Format(time.RFC3339),
	}

	responses.CacheControlCacheForWeek(w, r)

	err = s.tpl.ExecuteTemplate(w, "user.html", data)
	if err != nil {
		log.With("err", err).Error("failed to execute template")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func GetRouter(steamAPI *steamapi.SteamAPI, db *db.RedisStore) *chi.Mux {
	tpl, err := template.ParseFS(resources, "public/index.html", "public/user.html", "public/components.html")
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	s := &routeState{
		steamAPI: steamAPI,
		tpl:      tpl,
		db:       db,
	}

	r.Get("/", s.handleGetIndex)
	r.Post("/search", s.handlePostSearch)
	r.Get("/user/{steamid}", s.getUser)
	r.Get("/api/user/{steamid}", s.getUserAPI)

	r.Get("/lookup/{steamid}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("/user/%s", chi.URLParam(r, "steamid")), http.StatusSeeOther)
	})
	r.Get("/profiles/{steamid}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("/user/%s", chi.URLParam(r, "steamid")), http.StatusSeeOther)
	})

	r.Get("/id/{customurl}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("/?search=%s", chi.URLParam(r, "customurl")), http.StatusSeeOther)
	})

	fs, err := fs.Sub(resources, "public/static")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(fs))

	r.Handle("/favicon.ico", fileServer)
	r.Handle("/static/*", fileServer)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	return r
}
