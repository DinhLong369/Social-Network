package model

type User struct {
	Model
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"-"`
	Verified bool   `json:"verified,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}
