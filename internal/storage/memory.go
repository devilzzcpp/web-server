package storage

import (
	"sync"
	"web-server/internal/model"
	"web-server/pkg/logger"
)

type Storage struct {
	mu     sync.Mutex
	users  map[int]model.User
	nextID int
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
	m.mu.Lock()
	defer m.mu.Unlock()

	list := make([]model.User, 0, len(m.users))
	for _, user := range m.users {
		list = append(list, user)
	}
	return list
}

func (m *Storage) GetUser(id int) (model.User, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[id]
	return user, ok
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
	m.mu.Lock()
	defer m.mu.Unlock()

	list := make([]model.User, 0)
	for _, user := range m.users {
		if user.Role == role {
			list = append(list, user)
		}
	}
	return list
}

func (m *Storage) SeedUsers() {
	m.CreateUser(model.User{Username: "Леха", Role: "admin"})
	m.CreateUser(model.User{Username: "Андрей", Role: "user"})
	m.CreateUser(model.User{Username: "xd", Role: "moderator"})
}
