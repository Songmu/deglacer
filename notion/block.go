package notion

import "time"

type BlockList struct {
	Object     string      `json:"object"`
	Results    []Block     `json:"results"`
	NextCursor interface{} `json:"next_cursor"`
	HasMore    bool        `json:"has_more"`
}

type Block struct {
	BlockBase

	BulletedListItem BulletedListItemBlock `json:"bulleted_list_item"`
}

type BlockBase struct {
	Object         string    `json:"object"`
	Id             string    `json:"id"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_time"`
	HasChildren    bool      `json:"has_children"`
	Type           string    `json:"type"`
}

type BulletedListItemBlock struct {
	Text []RichText `json:"text"`
}

type Annotations struct {
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Code          bool   `json:"code"`
	Color         string `json:"color"`
}

type RichTextBase struct {
	PlainText   string      `json:"plain_text"`
	Href        string      `json:"href"`
	Type        string      `json:"type"`
	Annotations Annotations `json:"annotations"`
}

type RichText struct {
	RichTextBase

	Text struct {
		Content string `json:"content"`
		Link    struct {
			Type string `json:"type"`
			Url  string `json:"url"`
		} `json:"link"`
	} `json:"text"`
}
