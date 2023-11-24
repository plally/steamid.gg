package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/plally/steamid.id/internal/steamapi"
	"github.com/plally/steamid.id/internal/steamid"
)

//go:embed public/*
var resources embed.FS

type routeState struct {
	steamAPI *steamapi.SteamAPI
	tpl      *template.Template
}

func redirectError(w http.ResponseWriter, r *http.Request, err string) {
	err = url.QueryEscape(err)
	http.Redirect(w, r, fmt.Sprintf("/?error=%s", err), http.StatusSeeOther)
}

func (s *routeState) PostSearch(w http.ResponseWriter, r *http.Request) {
	queryString := r.FormValue("search")
	log := slog.With("query", queryString)

	q, err := parseSteamQuery(queryString)
	if err != nil {
		log.With("err", err).Info("failed to parse query")
		redirectError(w, r, "Could not find any steam user")
		return
	}

	if q.CustomURLName != "" {
		steamID, err := s.steamAPI.ResolveVanityURL(q.CustomURLName)
		if err != nil {
			log.With("err", err).Info("failed to resolve vanity url")
			redirectError(w, r, "Could not find any steam user")
			return
		}
		http.Redirect(w, r, "/user/"+steamID, http.StatusSeeOther)
	}

	if q.SteamID64 != "" {
		http.Redirect(w, r, "/user/"+q.SteamID64, http.StatusSeeOther)
	}

	if q.SteamID32 != "" {
		s, err := steamid.SteamID32(q.SteamID32)
		if err != nil {
			log.With("err", err).Info("failed to parse steamid32")
			redirectError(w, r, "Could not find any steam user")
			return
		}

		http.Redirect(w, r, "/user/"+s.SteamID64String(), http.StatusSeeOther)
	}
}

type IndexData struct {
	Error string
}

func (s *routeState) getIndex(w http.ResponseWriter, r *http.Request) {
	s.tpl.ExecuteTemplate(w, "index.html", IndexData{
		Error: r.URL.Query().Get("error"),
	})
}

type PlayerData struct {
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

func (s routeState) getUser(w http.ResponseWriter, r *http.Request) {
	steamID64 := chi.URLParam(r, "steamid")
	log := slog.With("steamid", steamID64)
	ply, err := s.steamAPI.GetPlayerSummary(steamID64)
	if err != nil {
		log.With("err", err).Error("failed to get player summary")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	steamID, err := steamid.SteamID64(steamID64)
	if err != nil {
		log.With("err", err).Error("failed to parse steamid64")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = s.tpl.ExecuteTemplate(w, "user.html", PlayerData{
		Username:    ply.PersonaName,
		Avatar:      ply.Avatarfull,
		CustomURL:   ply.ProfileURL,
		ProfileURL:  fmt.Sprintf("https://steamcommunity.com/profiles/%s", ply.Steamid),
		CreatedAt:   time.Unix(int64(ply.Timecreated), 0).Format(time.ANSIC),
		RealName:    ply.Realname,
		SteamID32:   steamID.SteamID32String(),
		SteamID64:   steamID.SteamID64String(),
		Location:    ply.Loccountrycode,
		SteamID3:    steamID.SteamID3String(),
		LastUpdated: time.Now().Format(time.ANSIC),
	})

	if err != nil {
		log.With("err", err).Error("failed to execute template")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

}

func GetRouter(steamAPI *steamapi.SteamAPI, tpl *template.Template) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	s := &routeState{
		steamAPI: steamAPI,
		tpl:      tpl,
	}

	r.Get("/", s.getIndex)
	r.Post("/search", s.PostSearch)
	r.Get("/user/{steamid}", s.getUser)

	fs, err := fs.Sub(resources, "public")
	if err != nil {
		panic(err)
	}

	r.Handle("/static/*", http.FileServer(http.FS(fs)))

	return r
}
