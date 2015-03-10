package web

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/julienschmidt/httprouter"
)

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data := commonSessionData(w, r)
	parseAndExecute(w, "index", data)
}

func topLangs(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := commonSessionData(w, r)
		data["langs"] = agg.PopularLanguages()
		data["lastrun"] = agg.LastRun().Local().Format("Jan 2, 2006 at 3:04pm")
		parseAndExecute(w, "toplangs", data)
	}
}

func topDevs(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := commonSessionData(w, r)
		data["devs"] = agg.PopularDevs()
		data["lastrun"] = agg.LastRun().Local().Format("Jan 2, 2006 at 3:04pm")
		parseAndExecute(w, "topdevs", data)
	}
}

func profile(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := commonSessionData(w, r)
		profile := agg.Profile(p.ByName("profile"))
		if profile != nil {
			data["profile"] = profile
			parseAndExecute(w, "profile", data)
		} else {
			parseAndExecute(w, "add", data)
		}
	}
}

func add(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := commonSessionData(w, r)
		user := data["user"]
		if user == nil {
			return
		}
		githubUser := user.(github.User)
		agg.Add(*githubUser.Login)
		http.Redirect(w, r, "/profile/"+*githubUser.Login, 302)
	}
}

func language(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := commonSessionData(w, r)
		data["languages"] = agg.Language(p.ByName("lang"))
		data["language"] = p.ByName("lang")
		parseAndExecute(w, "language", data)
	}
}

func admin(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := commonSessionData(w, r)
		if isAdmin, ok := data["admin"]; !ok || !isAdmin.(bool) {
			parseAndExecute(w, "403", data)
			return
		}
		data["lastRun"] = agg.LastRun().Local().Format("Jan 2, 2006 at 3:04pm")
		data["running"] = agg.Running()
		parseAndExecute(w, "admin", data)
	}
}

func adminCmd(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := commonSessionData(w, r)
		if isAdmin, ok := data["admin"]; !ok || !isAdmin.(bool) {
			parseAndExecute(w, "403", data)
			return
		}
		if r.FormValue("run") != "" {
			go agg.Run()
		}
		http.Redirect(w, r, "/admin", 302)
	}
}

func search(agg *aggregator.Aggregator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := commonSessionData(w, r)
		q := r.URL.Query().Get("q")
		data["q"] = q
		if q != "" {
			data["results"] = agg.Search(q)
		}
		parseAndExecute(w, "search", data)
	}
}
