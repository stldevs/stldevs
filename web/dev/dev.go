package dev

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/sessions"
	"log"
)

type ListQuery struct {
	Q    string `form:"q" binding:"required_without=Type"`
	Type string `form:"type" binding:"required_without=Q,omitempty,oneof=User Organization"`
}

func List(c *gin.Context) {
	var query ListQuery
	err := c.BindQuery(&query)
	if err != nil {
		log.Println(err)
		c.JSON(400, err.Error())
		return
	}

	if query.Q != "" {
		c.JSON(200, db.SearchUsers(query.Q))
		return
	}

	if listing := db.PopularDevs(query.Type); listing == nil {
		c.JSON(500, "Failed to list")
	} else {
		c.JSON(200, listing)
	}
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
	Hide bool `binding:"required"`
}

// Patch allows users and admins show or hide themselves in the site
func Patch(c *gin.Context) {
	login := c.Params.ByName("login")
	session := sessions.GetEntry(c)
	if session.User.IsAdmin == false && *session.User.Login != login {
		c.JSON(403, "Users can only modify themselves")
		return
	}

	profile, err := db.Profile(login)
	if err != nil || profile == nil {
		c.JSON(404, "Failed to find user")
		return
	}
	var cmd UpdateUser
	if err = c.BindJSON(&cmd); err != nil {
		c.JSON(400, "Failed to bind command object. Are you sending JSON?")
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
