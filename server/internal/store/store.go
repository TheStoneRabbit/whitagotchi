package store

import (
	"sync"

	"github.com/whitagotchi/whitagotchi/shared"
)

// Store is the persistence layer. For v0 this is an in-memory map;
// swap for SQLite once the API surface stabilizes.
type Store struct {
	mu        sync.RWMutex
	users     map[string]*User     // username -> user
	tokens    map[string]string    // token -> username
	creatures map[string]*shared.Creature // username -> creature
}

type User struct {
	Username string
	Token    string
}

func Open(_ string) (*Store, error) {
	return &Store{
		users:     map[string]*User{},
		tokens:    map[string]string{},
		creatures: map[string]*shared.Creature{},
	}, nil
}

func (s *Store) Close() error { return nil }

func (s *Store) CreateUser(u *User, c *shared.Creature) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.Username] = u
	s.tokens[u.Token] = u.Username
	s.creatures[u.Username] = c
	return nil
}

func (s *Store) UserByToken(token string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	name, ok := s.tokens[token]
	if !ok {
		return nil, false
	}
	return s.users[name], true
}

func (s *Store) UserByName(name string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[name]
	return u, ok
}

func (s *Store) Creature(username string) (*shared.Creature, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.creatures[username]
	return c, ok
}

func (s *Store) AllCreatures() []*shared.Creature {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*shared.Creature, 0, len(s.creatures))
	for _, c := range s.creatures {
		out = append(out, c)
	}
	return out
}

func (s *Store) UpdateCreature(c *shared.Creature) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.creatures[c.OwnerName] = c
}
