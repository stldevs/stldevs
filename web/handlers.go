package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func topLangs(cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		render(w, map[string]interface{}{
			"langs":   cmd.PopularLanguages(),
			"lastrun": cmd.LastRun(),
		})
	}
}

func topDevs(cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		render(w, map[string]interface{}{
			"devs":    cmd.PopularDevs(),
			"lastrun": cmd.LastRun(),
		})
	}
}

func topOrgs(cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		render(w, map[string]interface{}{
			"devs":    cmd.PopularOrgs(),
			"lastrun": cmd.LastRun(),
		})
	}
}

func profile(cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		profile, _ := cmd.Profile(p.ByName("profile"))
		render(w, map[string]interface{}{
			"profile": profile,
		})
	}
}

func language(cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		pageParam := r.URL.Query().Get("page")
		page := 0
		if pageParam != "" {
			var err error
			page, err = strconv.Atoi(pageParam)
			if err != nil {
				w.WriteHeader(400)
				return
			}
		}

		langs, userCount := cmd.Language(p.ByName("lang"), page)
		render(w, map[string]interface{}{
			"languages": langs,
			"count":     userCount,
			"language":  p.ByName("lang"),
			"page":      page,
		})
	}
}

func search(cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		q := r.URL.Query().Get("q")
		kind := r.URL.Query().Get("type")

		if q == "" {
			w.WriteHeader(400)
			return
		}

		render(w, map[string]interface{}{
			"results": cmd.Search(q, kind),
		})
	}
}

func render(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Error while rendering:", err)
		return
	}
}
