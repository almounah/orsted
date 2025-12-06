package userinfo

import (
	"os/user"
)

func GetUserName() string {
    u, err := user.Current()
	if err != nil {
		return "XXXXXXXX"
	}
    return u.Username
}

func GetUserIntegrity() string {
    return "TODO"
}
