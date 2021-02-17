package stldevs

import (
	"embed"
)

//go:embed swagger.json
var SwaggerDoc []byte

//go:embed swagger-ui/*
var SwaggerUI embed.FS
