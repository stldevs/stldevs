package web

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/julienschmidt/httprouter"
	"log"
	"encoding/json"
)

func topLangs(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := map[string]interface{}{}
		data["langs"] = cmd.PopularLanguages()
		if time, err := cmd.LastRun(); err == nil {
			data["lastrun"] = time
		}
		render(w, data)
	}
}

func topDevs(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := map[string]interface{}{}
		data["devs"] = cmd.PopularDevs()
		if time, err := cmd.LastRun(); err == nil {
			data["lastrun"] = time
		}
		render(w, data)
	}
}

func profile(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := map[string]interface{}{}
		profile, err := cmd.Profile(p.ByName("profile"))
		if err != nil {
			log.Println(err)
		}
		if profile != nil {
			data["profile"] = profile
		}
		render(w, data)
	}
}

func add(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		session := ctx.SessionData(w, r).Values
		user := session["user"]
		if user == nil {
			return
		}
		githubUser := user.(github.User)
		agg.Add(*githubUser.Login)
		http.Redirect(w, r, "/profile/"+*githubUser.Login, 302)
	}
}

func language(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := map[string]interface{}{}
		data["languages"] = cmd.Language(p.ByName("lang"))
		data["language"] = p.ByName("lang")
		render(w, data)
	}
}

func admin(ctx Context, cmd Commands, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := map[string]interface{}{}
		session := ctx.SessionData(w, r).Values
		if !isAdmin(session) {
			log.Println("User is not admin")
			return
		}
		if time, err := cmd.LastRun(); err == nil {
			data["lastRun"] = time
		}
		data["running"] = agg.Running()
		render(w, data)
	}
}

func adminCmd(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer func() {
			if recover := recover(); recover != nil {
				log.Println("Recovered in f", recover)
				http.Redirect(w, r, "/admin", 302)
			}
		}()
		session := ctx.SessionData(w, r).Values
		if !isAdmin(session) {
			return
		}
		if r.FormValue("run") != "" {
			go agg.Run()
		}
		http.Redirect(w, r, "/admin", 302)
	}
}

func search(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := map[string]interface{}{}
		data["session"] = ctx.SessionData(w, r).Values
		q := r.URL.Query().Get("q")
		data["q"] = q
		if q != "" {
			data["results"] = cmd.Search(q)
		}
		render(w, data)
	}
}

func isAdmin(s map[interface{}]interface{}) bool {
	if isAdmin, ok := s["admin"]; !ok || !isAdmin.(bool) {
		return false
	}
	return true
}

func render(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Error while rendering:", err)
		return
	}
}