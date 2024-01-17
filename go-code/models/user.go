package models

type User struct {
	ID        string
	Username  string
	Email     string
	pswdHash  string
	Active    string
	UserRole  string
}
