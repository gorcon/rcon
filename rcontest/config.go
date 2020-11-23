package rcontest

import "time"

// Config contains configuration for RCON Server.
// TODO: Implement me or remove.
type Config struct {
	Password             string
	AuthResponseDelay    time.Duration
	CommandResponseDelay time.Duration
}
