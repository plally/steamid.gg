package responses

import "net/http"

func CacheControlCacheForWeek(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "stale-while-revalidate=86400, public, max-age=604800")
	w.Header().Set("Vary", "Accept")

}

func CacheControlCache6Month(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "stale-while-revalidate=86400, public, max-age=15552000")
	w.Header().Set("Vary", "Accept")
}

func CacheControlNoCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Vary", "Accept")
}
