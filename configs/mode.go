package configs

import "fmt"

type Mode string

const (
	ModeCreate Mode = "create"
	ModeError  Mode = "error"
)

func parseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeCreate, ModeError:
		return Mode(s), nil
	default:
		return "", fmt.Errorf("invalid mode %q: must be one of [create, error]", s)
	}
}
