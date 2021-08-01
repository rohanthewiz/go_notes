package localsig

import "time"

type LocalSig struct {
	Id           int64
	Guid         string `sql:"size:40"`
	ServerSecret string `sql:"size:40"`
	CreatedAt    time.Time
}
