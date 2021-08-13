package peer

import (
	"errors"
	"go_notes/dbhandle"
	"time"
)

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

// We no longer create Peer here
// since peer needs to have been created to have an auth token
func GetPeerByGuid(peer_id string) (Peer, error) {
	var p Peer
	dbhandle.DB.Where("guid = ?", peer_id).First(&p)
	if p.Id < 1 {
		return p, errors.New("Could not create peer")
	}
	return p, nil
}
