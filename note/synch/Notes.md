
// VERIFICATION

/*	We did not follow some of the below
Synch philosophy - From the Perspective of the Client

    	- Have we met before? - Do I have you as a peer stored in my peer DB?
    		- Yes: Pull up the last synch point - Changeset from our Peer db
    			(the server should have a matching synch point for us)
    		- Else: Synch point will be 0 index of changesets sorted by created at ASC
    	- Are we in synch? - Does your latest change = my latest change?
    		Else let's synch

			Actual synching
			- From the synch point
				- Get all changesets from both sides more recent than the synch point
				- mark each change with a boolean 'local'
				- store in arr_unsynched_changes --> arr_uc
				- Sort by note guid, then by date asc
					- apply by desired algorithm
						(Thoughts)
						- apply by created_at asc?
						- don't apply changes I don't own
						- maintain the guid of the change, but it is recreated so new created_at on applyee
						(More Thoughts)
						- Apply update changes in date order however
							- Delete - Ends applying of changesets for that note
							- Create cannot follow Create or Update
							- In the DB make sure GUIDs are unique - so shouldn't have to check for create, update

			- We need to save our current synch point
				- sort the unsynched changes array by created_at
				- save the latest changeset guid in our peer db and the same guid in server's peer db
*/
