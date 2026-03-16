package db

import (
	"database/sql"
	"errors"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleGuest Role = "guest"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         Role
}

func UpsertUser(db *sql.DB, username, passwordHash string, role Role) error {
	_, err := db.Exec(`
		INSERT INTO users (username, password_hash, role)
		VALUES (?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET
			password_hash = excluded.password_hash,
			role          = excluded.role
	`, username, passwordHash, role)
	return err
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	row := db.QueryRow(`
		SELECT id, username, password_hash, role
		FROM users
		WHERE username = ?
	`, username)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
