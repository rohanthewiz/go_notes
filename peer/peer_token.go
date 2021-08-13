package peer

import (
	"errors"
	"go_notes/dbhandle"
	"go_notes/utils"
	"strings"
)

// Create Peer entry on server returning peer's auth token
// (If there is already a valid one for this peer, use that)
// This should be called on the server before first synch
// So the server will know of the peer and the token needed for access ahead of time
func GetPeerToken(peer_id string) (string, error) {
	var p Peer
	dbhandle.DB.Where("guid = ?", peer_id).First(&p)
	if p.Id < 1 {
		token := utils.GenerateSHA1()
		dbhandle.DB.Create(Peer{Guid: peer_id, Token: token})
		utils.Pl("Creating new peer entry for:", utils.ShortSHA(peer_id))
		dbhandle.DB.Where("guid = ?", peer_id).First(&p) // read it back
		if p.Id < 1 {
			return "", errors.New("Could not create peer entry")
		} else {
			return token, nil
		}
		// Peer already exists - make sure it has an auth token
	} else if len(p.Token) == 0 {
		token := utils.GenerateSHA1()
		p.Token = token
		dbhandle.DB.Save(&p)
		return token, nil
	} else {
		return p.Token, nil
	}
}

// The client will save the token for later access to the server
func SavePeerToken(compound string) {
	arr := strings.Split(strings.TrimSpace(compound), "-")
	peer_id, token := arr[0], arr[1]
	utils.Pf("Peer: %s, Auth Token: %s\n", peer_id, token)
	err := SetPeerToken(peer_id, token) // todo pull the error msg out of the err object
	if err != nil {
		utils.Pl(err)
	}
}

func SetPeerToken(peer_id string, token string) error {
	var p Peer
	dbhandle.DB.Where("guid = ?", peer_id).First(&p)
	if p.Id < 1 {
		utils.Pl("Creating new peer entry for:", utils.ShortSHA(peer_id))
		dbhandle.DB.Create(&Peer{Guid: peer_id, Token: token})
		// Verify
		dbhandle.DB.Where("guid = ?", peer_id).First(&p)
		if p.Id < 1 {
			return errors.New("Could not create peer entry")
		}
	} else { // Peer already exists - make sure it has an auth token
		p.Token = token // always update
		dbhandle.DB.Save(&p)
		utils.Pf("Updated token for peer entry: %s", utils.ShortSHA(peer_id))
	}
	return nil
}
