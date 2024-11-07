package models

type User struct {
	BaseItem     `json:",inline"`
	Name         string   `json:"name"`
	Roles        []string `json:"roles"`
	Campaigns    []string `json:"campaigns"`
	PasswordHash []byte   `json:"passwordhash"`
}

func (u User) GetName() string         { return u.Name }
func (u User) GetRoles() []string      { return u.Roles }
func (u User) GetCampaigns() []string  { return u.Campaigns }
func (u User) GetPasswordHash() []byte { return u.PasswordHash }
