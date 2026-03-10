package seeds

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/zalberix/cactus/apps/core/storage/db"
)

func init() {
	Register("users", SeedUsers)
}

func SeedUsers(ctx context.Context, storage *db.Queries) error {
	password := "password"
	hasher := sha512.New()
	hasher.Write([]byte(password))
	hashPassword := hex.EncodeToString(hasher.Sum(nil))

	_, err := storage.CreateUser(ctx, db.CreateUserParams{
		LastName:                "Demo",
		FirstName:               "User",
		Patronymic:              sql.NullString{Valid: true, String: "#1"},
		Email:                   "demo@mail.ru",
		Password:                hashPassword,
		ResetPasswordAfterLogin: sql.NullBool{Valid: true, Bool: false},
	})
	if err != nil {
		return fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return nil
}
