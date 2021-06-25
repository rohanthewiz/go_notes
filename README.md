## Go Notes - tracking your notes and snippets

This is a very fast note-taking and searching app written in Golang with both commandline and web interfaces.
No need to wait for a heavy GUI to load, just fire off go_notes with a few command line options and your tips and snippets are recorded to an SQLite database.
Markdown can be used for syntax highlighting in the body of the note.

## Getting Setup

### Download
Get Go for your operating system: http://golang.org/dl/ and install.
Make sure the full path to `<installation_folder>/bin` is in your PATH.
(The installation should set the GOROOT environment variable to the Golang toplevel folder.
If you installed Go to the folder ```~/apps/``` then your GOROOT env. var should contain: ```~/apps/go```).

### Go Workspace Setup
**Note:** this section is no longer necessary with go mods (you are using go mods, right? ;-) )

The environment variable *GOPATH* must point to your Go project workspace. Google 'set environment variable on Windows', for example, to do that for your operating system.
Your can check your current GOPATH and GOROOT environment variables by running the following command
```
go env
```
Your Go projects should live under a folder path of this format:
```
GOPATH/src/yourdomain.com/your_project
```

## Getting and Building GoNotes

```
# Clone the GoNotes archive
git clone https://github.com/rohanthewiz/go_notes.git
cd go_notes
go build # this will produce the executable 'go_notes' in the current directory
```

## Using GoNotes via Web Browser
GoNotes stores notes in four simple fields:
-   Title
-   Description
-   Body
-   Tags (Comma separated, list)

### Starting the Web Server

```
$ ./go_notes -svr
```

### Querying via Web Server
	/ql - Query for the last note updated
	/qi/:id - Query note by id
	/qg/:tag - Query note by tag
	/g/:tag - Query note by tag (a shorter form of the above)
	/qt/:title - Query note by title
	/q/:query - Query in tag, title, description or body
	/q/:query/l/:limit - Same as above, but with a limit
	/qg/:tag/q/:query - Query by tag and query term in any other field
	/g/:tag/:query - Same as above, but less typing
	/q/:query/qg/:tag - Same as above with just the order of query terms reversed
	/qt/:title/q/:query - Query by title and query term in any other field
	/q/:query/qt/:title - Same as above with just the order of query terms reversed

Examples:
```
    http://localhost:8092/ql
    http://localhost:8092/qi/5  # Show note with an index of 5
    http://localhost:8092/qg/todo # All notes with a tag of 'todo'
    http://localhost:8092/qt/cassandra # All notes with a title of 'cassandra'
    http://localhost:8092/q/test # All notes containing 'test' in any field
    http://localhost:8092/q/test/l/1 # Same as above but limit 1
    http://localhost:8092/qg/test/q/setup # All notes with a tag of 'test' and 'setup' in any other field
    http://localhost:8092/qt/cassandra/q/cluster # All notes with 'cassandra' in the title and 'cluster' in any other field
```

## Using GoNotes at Command Line

### Basic command line options

Creating a Note (quote option values with double-quotes if they contain spaces)

    -t Title
    -d Description
    -b Body
    -g "Comma separated, list, of, Tags"

Example:

```
./go_notes -t "My First Note" -d "Yep, it's my first note" -b "The body is where you give the long story about the note. I'm thinking you should be able to use all kinds of symbols. Double-quotes should be escaped with a backslash" -g "Test"
```

### Searching Notes (Command Line)
All searches are based on a LIKE (fuzzy) search
-q Query -- Retrieve notes based on a LIKE (fuzzy) search
-qi Integer -- Retrieve notes by ID (the number in square brackets on the left of the note is its ID)
-qg tag_to_search_for -- Retrieve notes by tag
-g tag_to_search_for -- Retrieve notes by tag (just a shorter form of the above)
-qg tag_to_search_for -q other_item -- Retrieve notes with tags that match tag_to_search_for and title, description or body that matches other_item
-qt title_to_search_for -q other_item -- Retrieve notes with title that match title_to_search_for and tag, description or body that matches other_item
-l number_of_notes -- Limit number of notes

Example:

```
$ ./go_notes -q note
[0] Title: My First Note - Yep, it's my first note
The body is where you give the long story about the note. I'm thinking you should be able to use all kinds of symbols. Double-quotes should be escaped with a backslash.
Tags: Test
```

### Updating (Command Line)
-upd -- Update an existing note - must be used with a query

Example:

```
$ ./go_notes -q "old note" -upd
```
#### Update Tips (Command Line)
* if a _+_ is placed at the start of an input field, that field of the current note will be _appended_ to
* if a _-_ is placed at the start of an input field, that field of the current note will be _blanked_
* if nothing is entered, the field will be unchanged

### Deleting (Command Line)
-del -- Delete an existing note - must be used with a query

Example:

```
$ ./go_notes -q trash -del
```

### Other Options
    
    -h -- List available options with defaults
    -db "" -- SQLite DB path. It will try to create the database 'go_notes.sqlite' in your home directory by default
    -l "-1" -- (formerly -ql) Limit the number of notes returned - default: -1 (no limit)
    -s Short Listing -- don't show the body
    -admin="" -- Privileged actions like 'delete_table' (drops the notes table)

Example:

```
D:\GoProjs\src\gotut.org\go_notes>go_notes -db "D:\xfr\gn.sqlite" -t "Test New DB Loc" -d "This is a test of the -db option"
$ gn -q all -s  # list all notes with the short list option. Note that here go_notes is aliased to *gn*
```

### Docker
- Build with `docker build --no-cache -t go_notes .`
- Run with `docker run --rm -p 8092:8092 go_notes:latest`
- Access the server at http://localhost:8092

### Synchronizing
Each GoNotes database is synchronizable with another. Just call one as the server and the other as a client.
The client and server must point to different databases whether remote or local.
```
# Remote - Hint get the server's IP first (`sudo ifconfig`)
$ ./go_notes -synch_server # Start the server
# Local
$ ./go_notes -synch_client server_address # Start the client. server_address is the IP address or name of server
```

Example - synch between two local databases
```
# Add a record to one database
$ ./go_notes -db db1.sqlite -t "A test note"  # Create a note in db1
$ ./go_notes -db db1.sqlite -get_server_secret # get a token for initial access to the server
$ ./go_notes -db db1.sqlite -synch_server
# Then in another terminal we run the client
# replace token_from_server below with the token received from the server
# The -server_secret option is only needed on the first access to the server from the client
$ ./go_notes -db db2.sqlite -synch_client localhost -server_secret token_from_server
$ ./go_notes -db db2.sqlite -q all # should now show the test note synched from db1
# Press CTRL-C on the server to quit
# delete db1.sqlite and db2.sqlite when complete
```

### TODO
- Token based auth webserver mode

### TIPS
- There is a great article on ego at http://blog.gopheracademy.com/advent-2014/ego/
- gcc is required to compile SQLite. On Windows you can get 64bit MinGW here http://mingw-w64.sourceforge.net/download.php. Install it and make sure to add the bin directory to your path
  (Could not get this to work with Windows 8.1, Windows 7 did work though)
- For less typing, you might want to do 'go build -o gn' (gn.exe on Windows) to produce the executable 'gn'
- This is a sweet way to learn a modern, highly performant language - The Go Programming Language using a database with an ORM (object relational manager) while building a useful tool!
I recommend using git to checkout an early version of go_notes so you can start out simple
- If you want to tinker with Go code and you are short on cast VSCode with the Go extension is your way to go. You ought to find a SQLite extension there also to peak into the SQLite file.
- If you are serious, and want to invest in first-class industrial strength tools, use JetBrains products. In this case Goland -- that is my tool of choice for Go, bar none.
- Feel free to create a pull request if you'd like to pitch in.

### Credits
- Go -- http://golang.org/  Thanks Google!!
- GORM -- https://github.com/jinzhu/gorm  - a decent abstraction over SQL.
- SQLite -- http://www.sqlite.org/ - A great place to start. SQLite is supported by GORM. 
