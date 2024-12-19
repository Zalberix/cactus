package schema

type Process struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
