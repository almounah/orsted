package userinfo

import (
	"os/user"
)

func GetUserName() string {
    u, _ := user.Current()
    return u.Username
}

func GetUserIntegrity() string {
    return "TODO"
}
