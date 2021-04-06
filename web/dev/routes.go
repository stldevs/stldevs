package dev

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/web/auth"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/devs",
	Handler:     List,
	Description: "List developers",
	Tags:        []string{"Developers"},
	Validate: crud.Validate{
		Query: crud.Object(map[string]crud.Field{
			"type": crud.String().Enum("User", "Organization").Description("List users or organizations. Required unless 'q' is provided."),
			"q":    crud.String().Description("Query string. Required unless 'type' is provided."),
		}),
	},
}, {
	Method:      "GET",
	Path:        "/devs/{login}",
	Handler:     Get,
	Description: "Get a developer by their GitHub username",
	Tags:        []string{"Developers"},
	Validate: crud.Validate{
		Path: crud.Object(map[string]crud.Field{
			"login": crud.String().Required(),
		}),
	},
}, {
	Method:      "PATCH",
	Path:        "/devs/{login}",
	PreHandlers: []gin.HandlerFunc{auth.Authenticated},
	Handler:     Patch,
	Description: "Allows users and admins show or hide themselves in the site",
	Tags:        []string{"Developers"},
	Validate: crud.Validate{
		Path: crud.Object(map[string]crud.Field{
			"login": crud.String().Required(),
		}),
		Body: crud.Object(map[string]crud.Field{
			"Hide": crud.Boolean().Required(),
		}),
	},
}, {
	Method:      "DELETE",
	Path:        "/devs/{login}",
	PreHandlers: []gin.HandlerFunc{auth.Authenticated},
	Handler:     Delete,
	Description: "Admins can expunge data until next run",
	Tags:        []string{"Developers"},
	Validate: crud.Validate{
		Path: crud.Object(map[string]crud.Field{
			"login": crud.String().Required(),
		}),
	},
}}
