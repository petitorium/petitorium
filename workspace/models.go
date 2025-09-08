package workspace

type Request struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	URL    string `yaml:"url"`
	Body   string `yaml:"body,omitempty"`
}

type Collection struct {
	Name        string       `yaml:"name"`
	Requests    []Request    `yaml:"requests,omitempty"`
	Collections []Collection `yaml:"collections,omitempty"`
	Expanded    bool         `yaml:"expanded,omitempty"`
}
