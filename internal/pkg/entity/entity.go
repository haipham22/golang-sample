package entity

var (
	SourceTypeNovel = "type"
	SourceTypeComic = "comic"
)

type Repository struct {
	Metadata Metadata `json:"metadata"`
	Sources  []Source `json:"data"`
}

type Metadata struct {
	Author      string `json:"author"`
	Description string `json:"description"`
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
