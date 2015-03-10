package web

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/julienschmidt/httprouter"
)

func index(ctx Context) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := ctx.SessionData(w, r)
		ctx.ParseAndExecute(w, "index", data)
	}
}

func topLangs(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := ctx.SessionData(w, r)
		data["langs"] = agg.PopularLanguages()
		data["lastrun"] = agg.LastRun().Local().Format("Jan 2, 2006 at 3:04pm")
		ctx.ParseAndExecute(w, "toplangs", data)
	}
}

func topDevs(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := ctx.SessionData(w, r)
		data["devs"] = agg.PopularDevs()
		data["lastrun"] = agg.LastRun().Local().Format("Jan 2, 2006 at 3:04pm")
		ctx.ParseAndExecute(w, "topdevs", data)
	}
}

func profile(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := ctx.SessionData(w, r)
		profile := agg.Profile(p.ByName("profile"))
		if profile != nil {
			data["profile"] = profile
			ctx.ParseAndExecute(w, "profile", data)
		} else {
			ctx.ParseAndExecute(w, "add", data)
		}
	}
}

func add(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := ctx.SessionData(w, r)
		user := data["user"]
		if user == nil {
			return
		}
		githubUser := user.(github.User)
		agg.Add(*githubUser.Login)
		http.Redirect(w, r, "/profile/"+*githubUser.Login, 302)
	}
}

func language(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := ctx.SessionData(w, r)
		data["languages"] = agg.Language(p.ByName("lang"))
		data["language"] = p.ByName("lang")
		ctx.ParseAndExecute(w, "language", data)
	}
}

func admin(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := ctx.SessionData(w, r)
		if isAdmin, ok := data["admin"]; !ok || !isAdmin.(bool) {
			ctx.ParseAndExecute(w, "403", data)
			return
		}
		data["lastRun"] = agg.LastRun().Local().Format("Jan 2, 2006 at 3:04pm")
		data["running"] = agg.Running()
		ctx.ParseAndExecute(w, "admin", data)
	}
}

func adminCmd(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := ctx.SessionData(w, r)
		if isAdmin, ok := data["admin"]; !ok || !isAdmin.(bool) {
			ctx.ParseAndExecute(w, "403", data)
			return
		}
		if r.FormValue("run") != "" {
			go agg.Run()
		}
		http.Redirect(w, r, "/admin", 302)
	}
}

func search(ctx Context, agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := ctx.SessionData(w, r)
		q := r.URL.Query().Get("q")
		data["q"] = q
		if q != "" {
			data["results"] = agg.Search(q)
		}
		ctx.ParseAndExecute(w, "search", data)
	}
}
