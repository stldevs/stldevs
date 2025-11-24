package web

import (
	"log"

	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/web/auth"
	"github.com/jakecoffman/stldevs/web/dev"
	"github.com/jakecoffman/stldevs/web/lang"
	"github.com/jakecoffman/stldevs/web/repo"
	"github.com/jakecoffman/stldevs/web/run"
)

func Run(cfg *config.Config) {
	r := crud.NewRouter("stldevs api", "1.0.0", crud.NewServeMuxAdapter())
	if cfg.Environment == "prod" {
		r.Swagger.BasePath = "/stldevs-api/"
	}

	must(r.Add(auth.New(cfg)...))
	must(r.Add(repo.Routes...))
	must(r.Add(run.Routes...))
	must(r.Add(dev.Routes...))
	must(r.Add(lang.Routes...))

	log.Println("Serving on http://127.0.0.1:8080")
	if err := r.Serve("0.0.0.0:8080"); err != nil {
		log.Println(err)
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
