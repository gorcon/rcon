package rcontest

import "time"

// Settings contains configuration for RCON Server.
type Settings struct {
	Password             string
	AuthResponseDelay    time.Duration
	CommandResponseDelay time.Duration
}

// Option allows to inject settings to Server.
type Option func(s *Server)

// SetSettings configuration for RCON Server.
func SetSettings(settings Settings) Option {
	return func(s *Server) {
		s.settings = settings
	}
}
