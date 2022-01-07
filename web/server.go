package web

import (
	"github.com/jakecoffman/crud"
	adapter "github.com/jakecoffman/crud/adapters/gin-adapter"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/web/auth"
	"github.com/jakecoffman/stldevs/web/dev"
	"github.com/jakecoffman/stldevs/web/lang"
	"github.com/jakecoffman/stldevs/web/repo"
	"github.com/jakecoffman/stldevs/web/run"
	"log"
)

func Run(cfg *config.Config) {
	r := crud.NewRouter("stldevs api", "1.0.0", adapter.New())
	if cfg.Environment == "prod" {
		r.Swagger.BasePath = "https://stldevs.com/stldevs-api/"
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
