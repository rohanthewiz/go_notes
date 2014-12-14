##Go Notes - Go, GORM and SQLite tracking your notes from the Command Line

This is a very fast command line note-taking and searching system.
No need to wait for a heavy GUI to load, just fire off go_notes with a few command line options and your tips and snippets are recorded to an SQLite database. This is also a sweet way to learn a modern, highly performant language - The Go Programming Language (aka Golang) using a database with an ORM (object relational manager) while building a useful tool!
Learning tip: use git to checkout an early version of go_notes so you can start out simple

##Getting Setup
###Download
 Get Go for your operating system: http://golang.org/dl/
###Go Workspace Setup
(Make sure Go is first downloaded)
The environment variable *GOPATH* must point to your Go project workspace. Google 'environment variable set Windows', for example, to do that for your operating system.
Your can check your current GOPATH by running the following command
```
go env
```

Your project should live under a folder path of this format:
```
GOPATH/src/yourdomain.com/your_go_project
```

##Building GoNotes
First clone the repo:
```
cd GOPATH/src/yourdomain.com
git clone https://github.com/rohanthewiz/go_notes.git
```

Build with

```
cd project_root_path/options
go get # should only need to run once
go install # build and install our options package
cd .. # project_root
go get # should only need to run once
go build go_notes.go
```

####TIPS
* gcc is required to compile SQLite. On Windows you can get 64bit MinGW here http://mingw-w64.sourceforge.net/download.php. Install it and make sure to add the bin directory to your path
* For less typing, you might want to alias the executable to 'gn' on your system. Google 'command alias' if neeeded.

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

###Other Options
    
    -h -- List available options with defaults
    -db "" -- Sqlite DB path. It will try to create the database 'go_notes.sqlite' in your home directory by default
    -qg "" -- Query by Tags column only
    -ql 9 -- Limit the number of notes returned
    -s Short Listing -- don't show the body
    -admin="" -- Privileged actions like 'delete_table' (drops the notes table)

Example:

```
D:\GoProjs\src\gotut.org\go_notes>go_notes -db "D:\xfr\gn.sqlite" -t "Test New DB Loc" -d "This is a test of the -db option"
$ gn -q all -s  # list all notes with the short list option. Note that here go_notes is aliased to *gn*
```

###TODO
- Import/export (csv, gob) is in the works
- Synching across a network
- webserver mode

###TIPS
Firefox has a great addon called SQLite Manager which you can use to peek into the database file
Feel free to create a pull request if you'd like to pitch in.

###Credits
- Go -- http://golang.org/  Thanks Google!!
- GORM -- https://github.com/jinzhu/gorm  - Who needs sluggish ActiveRecord, or other interpreted code interfacing to your database.
- SQLite -- http://www.sqlite.org/ - A great place to start. Actually GORM includes all the things needed for SQLite so SQLite gets compiled into GoNotes!
