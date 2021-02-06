package run

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/db"
	"time"
)

var epoch time.Time

func List(c *gin.Context) {
	if lastRun := db.LastRun(); lastRun.Year() == epoch.Year() {
		c.JSON(500, "Failed to list")
		return
	} else {
		c.JSON(200, lastRun)
	}
}
