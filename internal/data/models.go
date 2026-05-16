package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users       UsersModel
	Tokens      TokenModel
	Permissions PermissionModel
	Edocs       EdocModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:       UsersModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Edocs:       EdocModel{DB: db},
	}
}
