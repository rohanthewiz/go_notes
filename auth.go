// GOB client server
package main
import(
	"errors"
	"strings"
)

const authFailMsg = "Authentication failure. Generate authorization token with -synch_auth\nThen store in peer entry on client with -store_synch_auth"

// Get local DB signature
func whoAmI() string {
	var local_sig LocalSig
	db.First(&local_sig)
	if local_sig.Id < 1 {
		ensureDBSig()
		db.First(&local_sig)
		if local_sig.Id < 1 {
			fpl("Could not locate or create local database signature.\nYou should back up your notes, delete the local database, import your notes then try again")
			return ""
		}
	}
	return local_sig.Guid
}

func get_server_secret() string {
	var local_sig LocalSig
	db.First(&local_sig)
	if local_sig.Id > 0 {
		return local_sig.ServerSecret
	}
	return ""
}

// We no longer create Peer here
// since peer needs to have been created to have an auth token
func getPeerByGuid(peer_id string) (Peer, error) {
	var peer Peer
	db.Where("guid = ?", peer_id).First(&peer)
	if peer.Id < 1 {
		return peer, errors.New("Could not create peer")
	}
	return peer, nil
}

// Create Peer entry on server returning peer's auth token
// (If there is already a valid one for this peer, use that)
// This should be called on the server before first synch
// So the server will know of the peer and the token needed for access ahead of time
func getPeerToken(peer_id string) (string, error) {
	var peer Peer
	db.Where("guid = ?", peer_id).First(&peer)
	if peer.Id < 1 {
		token := generate_sha1()
		db.Create(&Peer{Guid: peer_id, Token: token})
		pl("Creating new peer entry for:", short_sha(peer_id))
		db.Where("guid = ?", peer_id).First(&peer) // read it back
		if peer.Id < 1 {
			return "", errors.New("Could not create peer entry")
		} else {
			return token, nil
		}
	  // Peer already exists - make sure it has an auth token
	} else if len(peer.Token) == 0 {
		token := generate_sha1()
		peer.Token = token
		db.Save(&peer)
		return token, nil
	} else {
		return peer.Token, nil
	}
}

// The client will save the token for later access to the server
func savePeerToken(compound string) {
	arr := strings.Split(strings.TrimSpace(compound), "-")
	peer_id, token := arr[0], arr[1]
	pf("Peer: %s, Auth Token: %s\n", peer_id, token)
	err := setPeerToken(peer_id, token)  // todo pull the error msg out of the err object
	if err != nil { pl(err) }
}

func setPeerToken(peer_id string, token string) (error) {
	var peer Peer
	db.Where("guid = ?", peer_id).First(&peer)
	if peer.Id < 1 {
		pl("Creating new peer entry for:", short_sha(peer_id))
		db.Create(&Peer{Guid: peer_id, Token: token})
		// Verify
		db.Where("guid = ?", peer_id).First(&peer)
		if peer.Id < 1 {
			return errors.New("Could not create peer entry")
		}
	} else { // Peer already exists - make sure it has an auth token
		peer.Token = token // always update
		db.Save(&peer)
		pf("Updated token for peer entry: %s", short_sha(peer_id))
	}
	return nil
}
