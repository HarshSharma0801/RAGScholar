package structure

import "encoding/xml"

type Link struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type Author struct {
	Name string `xml:"name"`
}

type Category struct {
	Term string `xml:"term,attr"`
}

type Entry struct {
	ID         string     `xml:"id"`
	Updated    string     `xml:"updated"`
	Published  string     `xml:"published"`
	Title      string     `xml:"title"`
	Summary    string     `xml:"summary"`
	Authors    []Author   `xml:"author"`
	Comment    string     `xml:"http://arxiv.org/schemas/atom comment"`
	Links      []Link     `xml:"link"`
	Categories []Category `xml:"category"`
	DOI        string     `xml:"http://arxiv.org/schemas/atom doi"`
	JournalRef string     `xml:"http://arxiv.org/schemas/atom journal_ref"`
}

type Feed struct {
	XMLName      xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Title        string   `xml:"title"`
	ID           string   `xml:"id"`
	Updated      string   `xml:"updated"`
	TotalResults int      `xml:"http://a9.com/-/spec/opensearch/1.1/ totalResults"`
	StartIndex   int      `xml:"http://a9.com/-/spec/opensearch/1.1/ startIndex"`
	ItemsPerPage int      `xml:"http://a9.com/-/spec/opensearch/1.1/ itemsPerPage"`
	Entries      []Entry  `xml:"entry"`
}

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
	Score           float32  `json:"score,omitempty"`
}

var Topics = []string{
	// Computer Science
	"machine learning", "cat:cs.LG",
	"artificial intelligence", "cat:cs.AI",
	"deep learning", "cat:cs.NE",
	"computer vision", "cat:cs.CV",
	"natural language processing", "cat:cs.CL",
	"data mining", "cat:cs.IR",
	"cryptography", "cat:cs.CR",
	"algorithms", "cat:cs.DS",
	"distributed computing", "cat:cs.DC",
	"quantum computing", "cat:cs.ET",
	// Physics
	"quantum physics", "cat:quant-ph",
	"condensed matter", "cat:cond-mat",
	"astrophysics", "cat:astro-ph",
	"high energy physics", "cat:hep-th",
	"optics", "cat:physics.optics",
	"fluid dynamics", "cat:physics.flu-dyn",
	"plasma physics", "cat:physics.plasm-ph",
	// Mathematics
	"algebra", "cat:math.AG",
	"combinatorics", "cat:math.CO",
	"number theory", "cat:math.NT",
	"probability", "cat:math.PR",
	"graph theory", "cat:math.CO",
	"differential geometry", "cat:math.DG",
	"topology", "cat:math.GT",
	// Other Fields
	"bioinformatics", "cat:q-bio.BM",
	"neuroscience", "cat:q-bio.NC",
	"robotics", "cat:cs.RO",
	"game theory", "cat:cs.GT",
	"statistics", "cat:stat.ML",
	"optimization", "cat:math.OC",
	"signal processing", "cat:eess.SP",
	"networks", "cat:cs.SI",
	"economics", "cat:econ.EM",
	"climate modeling", "cat:physics.ao-ph",
}
