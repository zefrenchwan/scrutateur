package services

// UserInformation is the json data definition to define an user
type UserInformation struct {
	Username string `json:"name"`
	Password string `json:"password"`
}
