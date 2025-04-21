package structure


type SimplifiedEntry struct {
	ID              string   `json:"id"`
	Updated         string   `json:"updated"`
	Published       string   `json:"published"`
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	Authors         []Author `json:"authors"`
	Comment         string   `json:"comment"`
	Links           []Link   `json:"links"`
	PrimaryCategory string   `json:"primaryCategory"`
	Categories      []string `json:"categories"`
	DOI             string   `json:"doi"`
	JournalRef      string   `json:"journalRef"`
}

type Author struct {
	Name string `json:"name"`
}

type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
	Type string `json:"type"`
}