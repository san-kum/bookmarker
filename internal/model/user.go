package model

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"password" json:"-"` // Don't include in JSON output
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func NewUser(username, password string) *User {
	return &User{
		Username:  username,
		Password:  password,
		CreatedAt: time.Now(),
	}
}
