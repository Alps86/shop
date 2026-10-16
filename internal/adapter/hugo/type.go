package hugo

import (
	"github.com/Alps86/shop/internal/domain/value"
)

type Page struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Layout      string `json:"layout"`
	FrameKind   string `json:"frame_kind"`
	RootVersion uint64 `json:"root_version"`
	Canonical   string `json:"canonical"`
	Fragment    string `json:"fragment"`
	ID          string `json:"-"`
}

type Pages []Page

type source struct {
	kind        string
	id          string
	slug        string
	name        string
	description string
	version     uint64
	locale      value.Locale
}
