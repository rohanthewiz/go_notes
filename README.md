##Go Notes - Go, GORM and SQLite tracking your notes from the Command Line

This is a very fast command line note-taking and searching system.
No need to wait for a heavy GUI to load, just fire off go_notes with a few command line options and your tips and snippets are recorded to an SQLite database. This is also a sweet way to learn a modern, performant language - Go (aka Golang) using a database with an ORM (object relational manager) while building a useful tool!
Tip: you might be able to alias this to gn on your system. Google how to alias commands on your system if neeeded.

Build with (Make sure Go is downloaded and setup first - http://golang.org/dl/)

```
go get
go build go_notes.go
```

###Windows Tip
gcc is required to compile SQLite. You can get 64bit MinGW here http://mingw-w64.sourceforge.net/download.php. Install it the make sure to add the bin directory to your path

###Basic command line options

Creating a Note (quote option values with double-quotes if they contain spaces)

    -t Title
    -d Description
    -b Body

Example:
```
./go_notes -t "My First Note" -d "Yep, it's my first note" -b "The body is where you give the long story about the note. I'm thinking you should be able to use all kinds of symbols too, but those need to be tested."
```

###Retrieving Notes

-q Query - Retrieve notes based on a LIKE search
-qi Integer - Retrieve notes by index (the number in square brackets on the left when the note is listed)

Example:

```
$ ./go_notes -q note
[0] Title: My First Note - Yep, it's my first note
Body: The body is where you give the long story about the note. I'm thinking you should be able to use all kinds of symbols too, but those need to be tested.
```

###Updating
-upd Update an existing note. Must be used with a query

Example:

```
$ ./go_notes -q "old note" -upd
```

###Deleting
-del Delete an existing note. Must be used with a query

Example:

```
$ ./go_notes -q trash -del
```

###Advanced Options

    -db "" Sqlite DB path. It will try to create the database 'go_notes.sqlite' in your home directory by default

Example:

```
D:\> D:\GoProjs\src\gotut.org\go_notes>go_notes -db "D:\xfr\gn.sqlite" -t "Test New DB Loc" -d "This is a test of the -db option"
```
    -g="" Tags - Not yet enabled - but would be a good exercise for the reader
    -ql=9 Limit the number of notes returned
    -s Short Listing - don't show the body
    -admin="" Privileged actions like 'delete_table' (drops the notes table)
###TODO
Update and delete is now implemented. So now we have full CRUD!
~~Update and delete existing notes. For now an SQLite tool can be used for deleting.~~

###TIPS
Firefox has a great addon called SQLite Manager which you can use to peek into the database file
Feel free to create a pull request if you'd like to pitch in.

###Credits
- Go - http://golang.org/  Thanks Google!
- GORM - https://github.com/jinzhu/gorm  - Who needs sluggish ActiveRecord, or other interpreted code interfacing to your database.
- SQLite - http://www.sqlite.org/ - A great place to start. Actually GORM includes all the things need for SQLite!
