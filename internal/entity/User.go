package entity

type User struct {
	Model
	Name string `json:"name"`
}

func (u *User) TableName() string {
	return "user"
}
