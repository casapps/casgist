package main

import (
	"embed"
	"io/fs"
)

// Embed static web assets
//go:embed web/static/*
var embeddedStatic embed.FS

// Embed HTML templates
//go:embed web/templates/*
var embeddedTemplates embed.FS

// GetStaticAssets returns the embedded static assets
func GetStaticAssets() fs.FS {
	subFS, err := fs.Sub(embeddedStatic, "web/static")
	if err != nil {
		panic("Failed to create static assets filesystem: " + err.Error())
	}
	return subFS
}

// GetTemplateAssets returns the embedded templates
func GetTemplateAssets() fs.FS {
	subFS, err := fs.Sub(embeddedTemplates, "web/templates")
	if err != nil {
		panic("Failed to create templates filesystem: " + err.Error())
	}
	return subFS
}