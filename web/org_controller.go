package web

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type OrgController struct {
	db    *sqlx.DB
	store *SessionStore
}

func (d *OrgController) List(c *gin.Context) {
	if listing := PopularOrgs(d.db); listing == nil {
		c.JSON(500, "Failed to list")
		return
	} else {
		c.JSON(200, listing)
	}
}

func (d *OrgController) Get(c *gin.Context) {
	profile, err := Profile(d.db, c.Params.ByName("login"))
	if err != nil {
		c.JSON(404, "Failed to find org")
		return
	}
	c.JSON(200, profile)
}
