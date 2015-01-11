##Go Notes - Go, GORM and SQLite tracking your notes from the Command Line

This is a very fast command line note-taking and searching system (web interface is being incorporated).
No need to wait for a heavy GUI to load, just fire off go_notes with a few command line options and your tips and snippets are recorded to an SQLite database.

##Getting Setup
###Download
 Get Go for your operating system: http://golang.org/dl/ and install
###Go Workspace Setup
The environment variable *GOPATH* must point to your Go project workspace. Google 'environment variable set Windows', for example, to do that for your operating system.
Your can check your current GOPATH by running the following command
```
go env
```

Your Go projects should live under a folder path of this format:
```
GOPATH/src/yourdomain.com/your_project
```

##Getting and Building GoNotes
```
go get github.com/rohanthewiz/go_notes
cd $GOPATH/src/github.com/rohanthewiz/go_notes
go build # this will produce the executable 'go_notes' in the current directory
```
You will require the ego package to make changes to the template
Install it with
```
go get github.com/benbjohnson/ego/...
```
If you make changes to an ego template file, you will need to 'compile' it before the main build.
For example if you updated *query.ego* you will need to run
```
$GOPATH/bin/ego -package main  # compile the template before doing 'go build'
```

##Using GoNotes

###Basic command line options

Creating a Note (quote option values with double-quotes if they contain spaces)

    -t Title
    -d Description
    -b Body
    -g "Comma, separated, list, of, Tags"

Example:
```
./go_notes -t "My First Note" -d "Yep, it's my first note" -b "The body is where you give the long story about the note. I'm thinking you should be able to use all kinds of symbols. Double-quotes should be escaped with a backslash" -g "Test"
```

###Retrieving Notes

-q Query -- Retrieve notes based on a LIKE search
-qi Integer -- Retrieve notes by ID (the number in square brackets on the left of the note is its ID)

Example:

```
$ ./go_notes -q note
[0] Title: My First Note - Yep, it's my first note
The body is where you give the long story about the note. I'm thinking you should be able to use all kinds of symbols. Double-quotes should be escaped with a backslash.
Tags: Test
```

###Updating
-upd -- Update an existing note - must be used with a query

Example:

```
$ ./go_notes -q "old note" -upd
```
####Update Tips
* if a _+_ is placed at the start of an input field, that field of the current note will be _appended_ to
* if a _-_ is placed at the start of an input field, that field of the current note will be _blanked_
* if nothing is entered, the field will be unchanged

###Deleting
-del -- Delete an existing note - must be used with a query

Example:

```
$ ./go_notes -q trash -del
```

###Web Server Mode
-svr -- Current only querying is implemented

Example:

```
$ ./go_notes -svr # then query for notes containing, for example 'todo' with localhost:8080/q/todo
```

###Other Options
    
    -h -- List available options with defaults
    -db "" -- Sqlite DB path. It will try to create the database 'go_notes.sqlite' in your home directory by default
    -qg "" -- Query by Tags column only
    -ql "-1" -- Limit the number of notes returned - default: -1 (no limit)
    -s Short Listing -- don't show the body
    -admin="" -- Privileged actions like 'delete_table' (drops the notes table)

Example:

```
D:\GoProjs\src\gotut.org\go_notes>go_notes -db "D:\xfr\gn.sqlite" -t "Test New DB Loc" -d "This is a test of the -db option"
$ gn -q all -s  # list all notes with the short list option. Note that here go_notes is aliased to *gn*
```

###TODO
- Finish up webserver mode
- Synching over a network

###TIPS
- There is a great article on ego at http://blog.gopheracademy.com/advent-2014/ego/
- gcc is required to compile SQLite. On Windows you can get 64bit MinGW here http://mingw-w64.sourceforge.net/download.php. Install it and make sure to add the bin directory to your path
  (Could not get this to work with Windows 8.1, Windows 7 did work though)
- For less typing, you might want to do 'go build -o gn' (gn.exe on Windows) to produce the executable 'gn'
- This is a sweet way to learn a modern, highly performant language - The Go Programming Language using a database with an ORM (object relational manager) while building a useful tool!
I recommend using git to checkout an early version of go_notes so you can start out simple
- Firefox has a great addon called SQLite Manager which you can use to peek into the database file
- Feel free to create a pull request if you'd like to pitch in.

###Credits
- Go -- http://golang.org/  Thanks Google!!
- GORM -- https://github.com/jinzhu/gorm  - Who needs sluggish ActiveRecord, or other interpreted code interfacing to your database.
- SQLite -- http://www.sqlite.org/ - A great place to start. Actually GORM includes all the things needed for SQLite so SQLite gets compiled into GoNotes!
