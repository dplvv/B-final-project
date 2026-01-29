package web

import "embed"

//go:embed templates/*.html templates/pages/*.html templates/partials/*.html static/*
var FS embed.FS
