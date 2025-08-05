package web

import "embed"

//go:embed template/*
var Templates embed.FS

//go:embed index.html
var Content embed.FS

