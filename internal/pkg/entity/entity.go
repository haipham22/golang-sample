package entity

var (
	SourceTypeNovel string = "type"
	SourceTypeComic string = "comic"
)

type Repository struct {
	Url     string   `json:"url"`
	Author  string   `json:"author"`
	Sources []Source `json:"data"`
}

type Source struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Path        string `json:"path"`
	Version     int    `json:"version"`
	Source      string `json:"source"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Locale      string `json:"locale"`
}
