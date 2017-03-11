package web

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type Handler func(context *Context, commands Commands) error

func mw(commands Commands, handlers ...Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := &Context{w, r, p}
		for _, h := range handlers {
			if err := h(ctx, commands); err != nil {
				break
			}
		}
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		log.Println(path, r.Method, r.RemoteAddr)
	}
}

func panicHandler(w http.ResponseWriter, _ *http.Request, d interface{}) {
	w.WriteHeader(500)
	if err := json.NewEncoder(w).Encode(d); err != nil {
		log.Println(err)
	}
}

func topLangs(ctx *Context, cmd Commands) error {
	return ctx.Render(map[string]interface{}{
		"langs":   cmd.PopularLanguages(),
		"lastrun": cmd.LastRun(),
	})
}

func topDevs(ctx *Context, cmd Commands) error {
	return ctx.Render(map[string]interface{}{
		"devs":    cmd.PopularDevs(),
		"lastrun": cmd.LastRun(),
	})
}

func topOrgs(ctx *Context, cmd Commands) error {
	return ctx.Render(map[string]interface{}{
		"devs":    cmd.PopularOrgs(),
		"lastrun": cmd.LastRun(),
	})
}

func profile(ctx *Context, cmd Commands) error {
	profile, _ := cmd.Profile(ctx.Params.ByName("profile"))
	return ctx.Render(map[string]interface{}{
		"profile": profile,
	})
}

func language(ctx *Context, cmd Commands) error {
	pageParam := ctx.Request.URL.Query().Get("page")
	page := 0
	if pageParam != "" {
		var err error
		page, err = strconv.Atoi(pageParam)
		if err != nil {
			ctx.Response.WriteHeader(400)
			return err
		}
	}

	langs, userCount := cmd.Language(ctx.Params.ByName("lang"), page)
	return ctx.Render(map[string]interface{}{
		"languages": langs,
		"count":     userCount,
		"language":  ctx.Params.ByName("lang"),
		"page":      page,
	})
}

func search(ctx *Context, cmd Commands) error {
	q := ctx.Request.URL.Query().Get("q")
	kind := ctx.Request.URL.Query().Get("type")

	if q == "" {
		ctx.Response.WriteHeader(400)
		return errors.New("q is empty")
	}

	return ctx.Render(map[string]interface{}{
		"results": cmd.Search(q, kind),
	})
}
