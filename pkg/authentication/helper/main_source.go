package helper

import "github.com/CHORUS-TRE/chorus-backend/internal/config"

func GetMainSourceID(cfg config.Config) string {
	for _, mode := range cfg.Services.AuthenticationService.Modes {
		if mode.MainSource {
			if mode.Type == "internal" {
				return "internal"
			}
			return mode.OpenID.ID
		}
	}
	return ""
}
