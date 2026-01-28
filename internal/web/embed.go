package web

import "embed"

//go:embed templates/**/*.html templates/*.html static/*
var FS embed.FS
