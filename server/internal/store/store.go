package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/whitagotchi/whitagotchi/shared"
)

// Store persists users and creatures in SQLite.
//
// Concurrency: we cap the connection pool to 1 and rely on WAL mode +
// busy_timeout so concurrent goroutines (engine tick, API handlers,
// chat hub) serialize cleanly without explicit locking.
type Store struct {
	db *sql.DB
}

type User struct {
	Username string
	Token    string
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
  username TEXT PRIMARY KEY,
  token    TEXT NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS creatures (
  owner_name   TEXT PRIMARY KEY REFERENCES users(username) ON DELETE CASCADE,
  id           TEXT NOT NULL,
  species      TEXT NOT NULL,
  rarity       TEXT NOT NULL,
  stage        TEXT NOT NULL,
  adult_form   TEXT NOT NULL DEFAULT '',
  quirk        TEXT NOT NULL,
  hunger       INTEGER NOT NULL,
  happiness    INTEGER NOT NULL,
  cleanliness  INTEGER NOT NULL,
  hatched_at   INTEGER NOT NULL,
  last_tick_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_token ON users(token);
`

func Open(path string) (*Store, error) {
	if path == "" {
		path = "whitagotchi.db"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, p := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
	} {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) CreateUser(u *User, c *shared.Creature) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO users(username, token) VALUES(?, ?)`, u.Username, u.Token); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO creatures(
		owner_name,id,species,rarity,stage,adult_form,quirk,
		hunger,happiness,cleanliness,hatched_at,last_tick_at
	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.OwnerName, c.ID,
		string(c.Species), string(c.Rarity), string(c.Stage), string(c.AdultForm), string(c.Quirk),
		c.Stats.Hunger, c.Stats.Happiness, c.Stats.Cleanliness,
		c.HatchedAt.UnixNano(), c.LastTickAt.UnixNano(),
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UserByToken(token string) (*User, bool) {
	return s.scanUser(s.db.QueryRow(`SELECT username, token FROM users WHERE token = ?`, token))
}

func (s *Store) UserByName(name string) (*User, bool) {
	return s.scanUser(s.db.QueryRow(`SELECT username, token FROM users WHERE username = ?`, name))
}

func (s *Store) scanUser(row *sql.Row) (*User, bool) {
	var u User
	if err := row.Scan(&u.Username, &u.Token); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// real error; treat as missing
		}
		return nil, false
	}
	return &u, true
}

const creatureCols = `owner_name,id,species,rarity,stage,adult_form,quirk,hunger,happiness,cleanliness,hatched_at,last_tick_at`

func (s *Store) Creature(username string) (*shared.Creature, bool) {
	row := s.db.QueryRow(`SELECT `+creatureCols+` FROM creatures WHERE owner_name = ?`, username)
	c, err := scanCreature(row.Scan)
	if err != nil {
		return nil, false
	}
	return c, true
}

func (s *Store) AllCreatures() []*shared.Creature {
	rows, err := s.db.Query(`SELECT ` + creatureCols + ` FROM creatures`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*shared.Creature
	for rows.Next() {
		c, err := scanCreature(rows.Scan)
		if err == nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *Store) UpdateCreature(c *shared.Creature) {
	_, _ = s.db.Exec(`UPDATE creatures SET
		id=?, species=?, rarity=?, stage=?, adult_form=?, quirk=?,
		hunger=?, happiness=?, cleanliness=?, hatched_at=?, last_tick_at=?
		WHERE owner_name=?`,
		c.ID,
		string(c.Species), string(c.Rarity), string(c.Stage), string(c.AdultForm), string(c.Quirk),
		c.Stats.Hunger, c.Stats.Happiness, c.Stats.Cleanliness,
		c.HatchedAt.UnixNano(), c.LastTickAt.UnixNano(),
		c.OwnerName,
	)
}

func scanCreature(scan func(...any) error) (*shared.Creature, error) {
	var (
		c                                 shared.Creature
		species, rarity, stage, adult, qk string
		hatched, lastTick                 int64
	)
	if err := scan(&c.OwnerName, &c.ID, &species, &rarity, &stage, &adult, &qk,
		&c.Stats.Hunger, &c.Stats.Happiness, &c.Stats.Cleanliness,
		&hatched, &lastTick,
	); err != nil {
		return nil, err
	}
	c.Species = shared.Species(species)
	c.Rarity = shared.Rarity(rarity)
	c.Stage = shared.Stage(stage)
	c.AdultForm = shared.AdultForm(adult)
	c.Quirk = shared.Quirk(qk)
	c.HatchedAt = time.Unix(0, hatched).UTC()
	c.LastTickAt = time.Unix(0, lastTick).UTC()
	return &c, nil
}
