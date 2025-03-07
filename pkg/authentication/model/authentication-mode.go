package model

type AuthenticationMode struct {
	Type       string
	Internal   Internal
	OpenID     OpenID
	ButtonText string
	IconURL    string
	Order      uint
}

type Internal struct {
	PublicRegistrationEnabled bool
}

type OpenID struct {
	ID string `yaml:"id"`
}
