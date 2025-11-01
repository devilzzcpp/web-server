package storage

import (
	"database/sql"
	"web-server/internal/model"
	"web-server/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

func (s *Storage) CreateUser(u model.User) (model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	u.Password = string(hashed)

	res, err := s.db.Exec(
		"INSERT INTO user (Username, Role, Login, Password) VALUES (?, ?, ?, ?)",
		u.Username, u.Role, u.Login, u.Password,
	)
	if err != nil {
		return model.User{}, err
	}

	id, _ := res.LastInsertId()
	u.ID = int(id)

	logger.Log.Info("создан пользователь", "id", u.ID, "username", u.Username)
	return u, nil
}

func (s *Storage) GetUser(id int) (model.User, bool, error) {
	var u model.User
	row := s.db.QueryRow("SELECT ID, Username, Role, Login, Password FROM user WHERE ID = ?", id)
	err := row.Scan(&u.ID, &u.Username, &u.Role, &u.Login, &u.Password)
	if err == sql.ErrNoRows {
		return model.User{}, false, nil
	} else if err != nil {
		return model.User{}, false, err
	}
	return u, true, nil
}

func (s *Storage) GetUsers() ([]model.User, error) {
	rows, err := s.db.Query("SELECT ID, Username, Role, Login, Password FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Login, &u.Password); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Storage) GetUsersByRole(role string) ([]model.User, error) {
	if role == "" {
		return s.GetUsers()
	}

	rows, err := s.db.Query(
		"SELECT ID, Username, Role, Login, Password FROM user WHERE Role = ?",
		role,
	)
	if err != nil {
		logger.Log.Error("ошибка запроса пользователей по роли", "role", role, "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Login, &u.Password); err != nil {
			logger.Log.Error("ошибка чтения строки пользователя", "error", err)
			return nil, err
		}
		users = append(users, u)
	}

	logger.Log.Info("получены пользователи по роли", "role", role, "count", len(users))
	return users, nil
}
