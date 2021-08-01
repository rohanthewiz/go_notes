package peer

import "time"

// A Peer represents a single client (db)
type Peer struct {
	Id        int64
	Guid      string `sql:"size:40"`
	Token     string `sql:"size:40"`
	User      string // (GUID) has one user //todo - Add Index
	Name      string `sql:"size:64"`
	SynchPos  string `sql:"size:40"` // Last changeset applied
	CreatedAt time.Time
	UpdatedAt time.Time
}
