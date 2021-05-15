package notion

import "time"

type Parent struct {
	Type       string `json:"type"`
	DatabaseID string `json:"database_id"`
	PageID     string `json:"page_id"`
	Workspace  bool   `json:"workspace"`
}

type Page struct {
	Object         string    `json:"object"`
	ID             string    `json:"id"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_time"`
	Parent         Parent    `json:"parent"`
	Archived       bool      `json:"archived"`
	Properties     struct {
		Title TitlePropertyValue `json:"title"`
	} `json:"properties"`
}

type TitlePropertyValue struct {
	Id    string     `json:"id"`
	Type  string     `json:"type"`
	Title []RichText `json:"title"`
}

func (p Page) PageTitle() string {
	if p.Parent.Type == "page" || p.Parent.Type == "workspace" {
		return p.Properties.Title.Title[0].PlainText
	}
	return ""
}
