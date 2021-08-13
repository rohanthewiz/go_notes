package authen

import (
	"fmt"
	"go_notes/dbhandle"
	"go_notes/localsig"
	"go_notes/migration"
)

// Get local DB signature
func WhoAmI() string {
	var local_sig localsig.LocalSig
	dbhandle.DB.First(&local_sig)
	if local_sig.Id < 1 {
		migration.EnsureDBSig()
		dbhandle.DB.First(&local_sig)
		if local_sig.Id < 1 {
			fmt.Println("Could not locate or create local database signature.\nYou should back up your notes, delete the local database, import your notes then try again")
			return ""
		}
	}
	return local_sig.Guid
}

func GetServerSecret() string {
	var local_sig localsig.LocalSig
	dbhandle.DB.First(&local_sig)
	if local_sig.Id > 0 {
		return local_sig.ServerSecret
	}
	return ""
}
