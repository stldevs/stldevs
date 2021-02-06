package org

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/db"
)

func List(c *gin.Context) {
	if listing := db.PopularOrgs(); listing == nil {
		c.JSON(500, "Failed to list")
		return
	} else {
		c.JSON(200, listing)
	}
}

func Get(c *gin.Context) {
	profile, err := db.Profile(c.Params.ByName("login"))
	if err != nil {
		c.JSON(404, "Failed to find org")
		return
	}
	c.JSON(200, profile)
}
