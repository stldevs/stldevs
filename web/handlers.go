package web

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func topLangs(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	c.JSON(200, map[string]interface{}{
		"langs":   PopularLanguages(db),
		"lastrun": LastRun(db),
	})
}

func topDevs(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	c.JSON(200, map[string]interface{}{
		"devs":    PopularDevs(db),
		"lastrun": LastRun(db),
	})
}

func topOrgs(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	c.JSON(200, map[string]interface{}{
		"devs":    PopularOrgs(db),
		"lastrun": LastRun(db),
	})
}

func profile(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	profile, _ := Profile(db, c.Params.ByName("profile"))
	c.JSON(200, map[string]interface{}{
		"profile": profile,
	})
}

func language(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	langs, userCount := Language(db, c.Params.ByName("lang"))
	c.JSON(200, map[string]interface{}{
		"languages": langs,
		"count":     userCount,
		"language":  c.Params.ByName("lang"),
	})
}

func search(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	q := c.Request.URL.Query().Get("q")
	kind := c.Request.URL.Query().Get("type")

	if q == "" {
		c.Status(400)
		return
	}

	c.JSON(200, map[string]interface{}{
		"results": Search(db, q, kind),
	})
}
