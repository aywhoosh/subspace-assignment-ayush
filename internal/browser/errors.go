package browser

import "errors"

var (
	ErrCheckpoint = errors.New("automation paused: checkpoint requires manual intervention")
)
