package generator

// Link represents a clickable link with display text and URL
type Link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// Achievement represents a single bullet point with optional overflow flag
type Achievement struct {
	Text     string `json:"text"`
	Overflow bool   `json:"overflow,omitempty"`
}

// Contact contains all contact information for the resume header
type Contact struct {
	Location string `json:"location"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Website  Link   `json:"website"`
	LinkedIn Link   `json:"linkedin"`
	GitHub   Link   `json:"github"`
}

// Skill represents a skill category with its items
type Skill struct {
	Category string `json:"category"`
	Items    string `json:"items"`
	Overflow bool   `json:"overflow,omitempty"`
}

// Experience represents a work experience entry
type Experience struct {
	Company      string        `json:"company"`
	Title        string        `json:"title"`
	URL          string        `json:"url"`
	Dates        string        `json:"dates"`
	Achievements []Achievement `json:"achievements"`
	Tech         string        `json:"tech"`
}

// Project represents a personal or professional project
type Project struct {
	Name         string        `json:"name"`
	URL          string        `json:"url"`
	Achievements []Achievement `json:"achievements"`
	Tech         string        `json:"tech"`
}

// Education represents educational background
type Education struct {
	School string `json:"school"`
	Degree string `json:"degree"`
	URL    string `json:"url"`
	Date   string `json:"date"`
}

// Resume is the complete resume data structure
type Resume struct {
	Name       string       `json:"name"`
	Title      string       `json:"title"`
	Contact    Contact      `json:"contact"`
	Summary    string       `json:"summary"`
	Skills     []Skill      `json:"skills"`
	Experience []Experience `json:"experience"`
	Projects   []Project    `json:"projects"`
	Education  []Education  `json:"education"`
}
