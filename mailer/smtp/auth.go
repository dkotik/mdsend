package smtp

import (
	"errors"
	"net/smtp"
	"strings"
)

// https://gist.github.com/jpillora/cb46d183eca0710d909a
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) (smtp.Auth, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return nil, errors.New("SMTP: username and password are required")
	}
	return &loginAuth{username, password}, nil
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("uknown server request")
		}
	}
	return nil, nil
}
