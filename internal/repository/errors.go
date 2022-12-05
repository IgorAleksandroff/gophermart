package repository

import "errors"

var ErrUserRegister = errors.New("user already exist")
var ErrUserLogin = errors.New("unknown user")
