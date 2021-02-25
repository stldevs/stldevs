package run

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/db"
	"time"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/runs",
	Handler:     List,
	Description: "Gets the time of the last scrape of GitHub",
	Tags:        []string{"Last Run"},
	Validate:    crud.Validate{},
}}

var epoch time.Time

func List(c *gin.Context) {
	if lastRun := db.LastRun(); lastRun.Year() == epoch.Year() {
		c.JSON(500, "Failed to list")
		return
	} else {
		c.JSON(200, lastRun)
	}
}
