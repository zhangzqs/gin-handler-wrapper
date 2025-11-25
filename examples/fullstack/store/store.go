package store

import (
	"errors"
	"sync"
	"time"

	"github.com/zhangzqs/go-typed-rpc/examples/fullstack/model"
)

// ==================== 模拟数据存储层 ====================

// Store 数据存储
type Store struct {
	users    map[int64]model.User
	articles map[int64]model.Article
	nextID   int64
	mu       sync.RWMutex
}

var (
	instance *Store
	once     sync.Once
)

// GetStore 获取存储单例
func GetStore() *Store {
	once.Do(func() {
		instance = &Store{
			users:    make(map[int64]model.User),
			articles: make(map[int64]model.Article),
			nextID:   1,
		}
		// 初始化测试数据
		instance.initTestData()
	})
	return instance
}

// initTestData 初始化测试数据
func (s *Store) initTestData() {
	s.users[1] = model.User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
	}
	s.users[2] = model.User{
		ID:        2,
		Name:      "Bob",
		Email:     "bob@example.com",
		CreatedAt: time.Now(),
	}
	s.nextID = 3
}

// CreateUser 创建用户
func (s *Store) CreateUser(name, email string) model.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := model.User{
		ID:        s.nextID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}
	s.users[s.nextID] = user
	s.nextID++
	return user
}

// GetUser 获取用户
func (s *Store) GetUser(id int64) (model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return model.User{}, errors.New("user not found")
	}
	return user, nil
}

// ListUsers 获取用户列表
func (s *Store) ListUsers() []model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userList := make([]model.User, 0, len(s.users))
	for _, user := range s.users {
		userList = append(userList, user)
	}
	return userList
}

// DeleteUser 删除用户
func (s *Store) DeleteUser(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}
	delete(s.users, id)
	return nil
}

// UpdateArticle 更新文章
func (s *Store) UpdateArticle(id int64, title, content string) model.Article {
	s.mu.Lock()
	defer s.mu.Unlock()

	article, exists := s.articles[id]
	if !exists {
		// 如果不存在就创建
		article = model.Article{
			ID:     id,
			Author: "unknown",
		}
	}
	article.Title = title
	article.Content = content
	s.articles[id] = article
	return article
}
