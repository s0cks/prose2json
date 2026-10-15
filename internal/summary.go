package internal

type BookSummary struct {
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	KeyThemes []string `json:"key_themes"`
	WordCount int      `json:"word_count"`
}
