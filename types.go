package main

type user struct {
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}

type comment struct {
	Body   string `json:"body"`
	Author user   `json:"author"`
}

type issue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Reporter user `json:"reporter"`
	} `json:"fields"`
}
