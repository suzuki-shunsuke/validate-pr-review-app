package github

import (
	"strings"

	v4 "github.com/suzuki-shunsuke/require-pr-review-app/pkg/github/v4"
)

type User struct {
	Login string `json:"login"`
	IsApp bool   `json:"is_app"`
}

func newUser(v *v4.User) *User {
	return &User{
		Login: v.Login,
		IsApp: strings.HasPrefix(v.ResourcePath, "/apps/") || strings.HasSuffix(v.Login, "[bot]"),
	}
}
