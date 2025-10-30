package storage

import (
	"sync"
	"web-server/internal/model"
	"web-server/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	mu     sync.Mutex
	users  map[int]model.User
	nextID int
}

func (m *Storage) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (m *Storage) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func NewStorage() *Storage {
	return &Storage{
		users:  make(map[int]model.User),
		nextID: 1,
	}
}

func (m *Storage) CreateUser(user model.User) model.User {
	m.mu.Lock()
	defer m.mu.Unlock()

	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user

	logger.Log.Info("создан новый пользователь",
		"id", user.ID,
		"username", user.Username,
		"role", user.Role,
	)

	return user
}

func (m *Storage) GetUsers() []model.User {
	list := make([]model.User, 0, len(m.users))
	for _, user := range m.users {
		list = append(list, user)
	}
	return list
}

func (m *Storage) GetUser(id int) (model.User, bool) {
	user, ok := m.users[id]
	return user, ok
}

func (m *Storage) GetUserByLogin(login string) (model.User, bool) {
	for _, user := range m.users {
		if user.Login == login {
			return user, true
		}
	}
	return model.User{}, false
}

func (m *Storage) UpdateUser(id int, user model.User) (model.User, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.users[id]
	if !ok {
		return model.User{}, false
	}
	user.ID = existing.ID
	m.users[id] = user
	logger.Log.Info("обновлен пользователь",
		"id", id,
		"username", user.Username,
		"role", user.Role,
	)
	return user, true
}

func (m *Storage) DeleteUser(id int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.users[id]; !ok {
		return false
	}
	delete(m.users, id)
	logger.Log.Info("удален пользователь", "id", id)
	return true
}

func (m *Storage) GetUsersByRole(role string) []model.User {
	list := make([]model.User, 0)
	for _, user := range m.users {
		if user.Role == role {
			list = append(list, user)
		}
	}
	return list
}

func (m *Storage) SeedUsers() {
	users := []model.User{
		{Username: "Леха", Role: "admin", Login: "admin", Password: "12345"},
		{Username: "Андрей", Role: "user", Login: "andrey", Password: "12345"},
		{Username: "xd", Role: "moderator", Login: "xd", Password: "12345"},
	}

	for _, u := range users {
		// Хэшируем пароль перед созданием
		hashed, err := m.HashPassword(u.Password)
		if err != nil {
			logger.Log.Error("ошибка хэширования пароля", "user", u.Username, "error", err)
			continue
		}
		u.Password = hashed
		m.CreateUser(u)
	}
}
