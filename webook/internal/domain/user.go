package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Phone    string
	Password string
	Nickname string
	Birthday string
	AboutMe  string
	Ctime    time.Time
}
