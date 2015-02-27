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
			println("Could not locate or create local database signature.\nYou should back up your notes, delete the local database, import your notes then try again")
			return ""
		}
	}
	return local_sig.Guid
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
// This should be called on the server before first synch
// So the server will know of the peer and the token needed for access ahead of time
func getPeerToken(peer_id string) (string, error) {
	var peer Peer
	db.Where("guid = ?", peer_id).First(&peer)
	if peer.Id < 1 {
		token := generate_sha1()
		db.Create(&Peer{Guid: peer_id, Token: token})
		println("Creating new peer entry for:", short_sha(peer_id))
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
func savePeerToken(compound string) (error) {
//	var peer_id string
//	var token string
	// Split compund string
	arr := strings.Split(compound, "-")
	peer_id := arr[0]; token := arr[1]

	var peer Peer
	db.Where("guid = ?", peer_id).First(&peer)
	if peer.Id < 1 { // then Create
//		token := generate_sha1()
		db.Create(&Peer{Guid: peer_id, Token: token})
		println("Creating new peer entry for:", short_sha(peer_id))
		db.Where("guid = ?", peer_id).First(&peer) // read it back
		if peer.Id < 1 {
			return errors.New("Could not create peer entry")
		}
	  // Peer already exists - make sure it has an auth token
	} else if len(peer.Token) == 0 {
		peer.Token = token
		db.Save(&peer)
	}

	return nil
}
