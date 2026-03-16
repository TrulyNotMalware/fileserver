package db

import (
	"database/sql"
	"errors"
	"time"
)

func SaveRefreshToken(db *sql.DB, userID int64, token string, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, userID, token, expiresAt.UTC())
	return err
}

func GetRefreshToken(db *sql.DB, token string) (userID int64, expiresAt time.Time, err error) {
	row := db.QueryRow(`
		SELECT user_id, expires_at
		FROM refresh_tokens
		WHERE token = ?
	`, token)

	if err = row.Scan(&userID, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, time.Time{}, nil
		}
		return 0, time.Time{}, err
	}

	return userID, expiresAt, nil
}

func DeleteRefreshToken(db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE token = ?`, token)
	return err
}

func DeleteExpiredRefreshTokens(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < ?`, time.Now().UTC())
	return err
}
