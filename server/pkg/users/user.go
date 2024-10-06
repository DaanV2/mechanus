package users

type User struct {
	ID           string   `json:"id"`
	Roles        []string `json:"roles"`
	Campaigns    []string `json:"campaigns"`
	PasswordHash []byte   `json:"passwordhash"`
}
