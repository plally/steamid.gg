package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/plally/steamid.id/internal/steamapi"
)

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
		http.Redirect(w, r, "/user/"+steamID, http.StatusFound)
	}

	if q.SteamID64 != "" {
		http.Redirect(w, r, "/user/"+q.SteamID64, http.StatusFound)
	}

	if q.SteamID32 != "" {
		steamID64, err := steamID32to64(q.SteamID32)
		if err != nil {
			log.With("err", err).Info("failed to convert steamid32 to steamid64")
			redirectError(w, r, "Could not find any steam user")
			return
		}
		http.Redirect(w, r, "/user/"+steamID64, http.StatusFound)
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
	Username   string
	Avatar     string
	CustomURL  string
	ProfileURL string
	RealName   string
	SteamID32  string
	SteamID64  string
	Location   string
	CreatedAt  string
}

func (s routeState) getUser(w http.ResponseWriter, r *http.Request) {
	steamID := chi.URLParam(r, "steamid")
	log := slog.With("steamid", steamID)
	ply, err := s.steamAPI.GetPlayerSummary(steamID)
	if err != nil {
		log.With("err", err).Error("failed to get player summary")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	steam32, err := steamID64to32(steamID)
	if err != nil {
		log.With("err", err).Error("failed to convert steamid64 to steamid32")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = s.tpl.ExecuteTemplate(w, "user.html", PlayerData{
		Username:   ply.PersonaName,
		Avatar:     ply.Avatarfull,
		CustomURL:  ply.ProfileURL,
		ProfileURL: fmt.Sprintf("https://steamcommunity.com/profiles/%s", ply.Steamid),
		CreatedAt:  time.Unix(int64(ply.Timecreated), 0).Format(time.UnixDate),
		RealName:   ply.Realname,
		SteamID32:  steam32,
		SteamID64:  ply.Steamid,
		Location:   ply.Loccountrycode,
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

	return r
}
