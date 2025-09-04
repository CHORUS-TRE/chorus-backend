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

const DEFAULT_USERNAME_CLAIM = "sub"
const DEFAULT_FIRST_NAME_CLAIM = "given_name"
const DEFAULT_LAST_NAME_CLAIM = "family_name"
