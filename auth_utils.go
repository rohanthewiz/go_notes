// GOB client server
package main

import (
	"errors"
	"fmt"
	db "go_notes/dbhandle"
	"go_notes/localsig"
	"go_notes/migration"
	"go_notes/peer"
	"go_notes/utils"
	"strings"
)

const authFailMsg = "Authentication failure. Generate authorization token with -synch_auth\nThen store in peer entry on client with -store_synch_auth"

// Get local DB signature
func whoAmI() string {
	var local_sig localsig.LocalSig
	db.DB.First(&local_sig)
	if local_sig.Id < 1 {
		migration.EnsureDBSig()
		db.DB.First(&local_sig)
		if local_sig.Id < 1 {
			fmt.Println("Could not locate or create local database signature.\nYou should back up your notes, delete the local database, import your notes then try again")
			return ""
		}
	}
	return local_sig.Guid
}

func getServerSecret() string {
	var local_sig localsig.LocalSig
	db.DB.First(&local_sig)
	if local_sig.Id > 0 {
		return local_sig.ServerSecret
	}
	return ""
}

// We no longer create Peer here
// since peer needs to have been created to have an auth token
func getPeerByGuid(peer_id string) (peer.Peer, error) {
	var p peer.Peer
	db.DB.Where("guid = ?", peer_id).First(&p)
	if p.Id < 1 {
		return p, errors.New("Could not create peer")
	}
	return p, nil
}

// Create Peer entry on server returning peer's auth token
// (If there is already a valid one for this peer, use that)
// This should be called on the server before first synch
// So the server will know of the peer and the token needed for access ahead of time
func getPeerToken(peer_id string) (string, error) {
	var p peer.Peer
	db.DB.Where("guid = ?", peer_id).First(&p)
	if p.Id < 1 {
		token := utils.GenerateSHA1()
		db.DB.Create(peer.Peer{Guid: peer_id, Token: token})
		utils.Pl("Creating new peer entry for:", utils.ShortSHA(peer_id))
		db.DB.Where("guid = ?", peer_id).First(&p) // read it back
		if p.Id < 1 {
			return "", errors.New("Could not create peer entry")
		} else {
			return token, nil
		}
		// Peer already exists - make sure it has an auth token
	} else if len(p.Token) == 0 {
		token := utils.GenerateSHA1()
		p.Token = token
		db.DB.Save(&p)
		return token, nil
	} else {
		return p.Token, nil
	}
}

// The client will save the token for later access to the server
func savePeerToken(compound string) {
	arr := strings.Split(strings.TrimSpace(compound), "-")
	peer_id, token := arr[0], arr[1]
	utils.Pf("Peer: %s, Auth Token: %s\n", peer_id, token)
	err := setPeerToken(peer_id, token) // todo pull the error msg out of the err object
	if err != nil {
		utils.Pl(err)
	}
}

func setPeerToken(peer_id string, token string) error {
	var p peer.Peer
	db.DB.Where("guid = ?", peer_id).First(&p)
	if p.Id < 1 {
		utils.Pl("Creating new peer entry for:", utils.ShortSHA(peer_id))
		db.DB.Create(&peer.Peer{Guid: peer_id, Token: token})
		// Verify
		db.DB.Where("guid = ?", peer_id).First(&p)
		if p.Id < 1 {
			return errors.New("Could not create peer entry")
		}
	} else { // Peer already exists - make sure it has an auth token
		p.Token = token // always update
		db.DB.Save(&p)
		utils.Pf("Updated token for peer entry: %s", utils.ShortSHA(peer_id))
	}
	return nil
}
