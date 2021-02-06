package dev

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/sessions"
)

func List(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		if listing := db.PopularDevs(); listing == nil {
			c.JSON(500, "Failed to list")
		} else {
			c.JSON(200, listing)
		}
		return
	}

	c.JSON(200, db.SearchUsers(q))
}

func Get(c *gin.Context) {
	profile, err := db.Profile(c.Params.ByName("login"))
	if err != nil {
		c.JSON(404, "Failed to find user")
		return
	}
	c.JSON(200, profile)
}

type UpdateUser struct {
	Hide bool
}

// Patch allows users and admins show or hide themselves in the site
func Patch(c *gin.Context) {
	profile, err := db.Profile(c.Params.ByName("login"))
	if err != nil {
		c.JSON(404, "Failed to find user")
		return
	}
	var cmd UpdateUser
	if err = c.BindJSON(&cmd); err != nil {
		c.JSON(400, "Failed to bind command object. Are you sending JSON?")
		return
	}
	session := sessions.GetEntry(c)
	if session.User.IsAdmin == false && *session.User.Login != *profile.User.Login {
		c.JSON(403, "Users can only modify themselves")
		return
	}
	err = db.HideUser(cmd.Hide, *profile.User.Login)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, profile)
}

// Delete allows admins to easily expunge old data
func Delete(c *gin.Context) {
	session := sessions.GetEntry(c)
	if session.User.IsAdmin == false {
		c.JSON(403, "Only admins can delete users")
		return
	}

	login := c.Params.ByName("login")

	err := db.Delete(login)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, "deleted")
}
