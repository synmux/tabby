// Package web provides embedded static assets and HTML templates for Tabby.
package web

import "embed"

// StaticFiles contains all files under static/ (CSS, JS, images).
//
//go:embed static
var StaticFiles embed.FS

// TemplateFiles contains all HTML templates.
//
//go:embed templates
var TemplateFiles embed.FS
