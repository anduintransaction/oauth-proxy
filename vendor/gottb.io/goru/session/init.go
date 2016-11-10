package session

import (
	"gottb.io/goru"
	"gottb.io/goru/config"
)

func Start(config *config.Config) error {
	goru.SetSessionStore(&cookieStore{})
	return nil
}
