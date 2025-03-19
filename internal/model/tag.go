package model

type Tag struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func NewTag(name string) Tag {
	return Tag{
		Name: name,
	}
}
