Scribble (FireScribble Edition) [![GoDoc](https://godoc.org/github.com/boltdb/bolt?status.svg)](http://godoc.org/github.com/creativeguy2013/scribble) [![Go Report Card](https://goreportcard.com/badge/github.com/creativeguy2013/scribble)](https://goreportcard.com/report/github.com/creativeguy2013/scribble)
--------

A tiny JSON database in Golang - behaviour is very similar to Google Cloud Firestore


### Installation

Install using `go get github.com/creativeguy2013/scribble`.

### Usage

```go
// a new scribble driver, providing the directory where it will be writing to
db, err := scribble.New(dir)
if err != nil {
  fmt.Println("Error", err)
}

// open a collection from the base document
fishCollection := db.Collection("fish")

// open the document we want to write to
onefishDocument := fishCollection.Document("onefish")

// write the data to the document
fish := Fish{}
if err := onefishDocument.Write(fish); err != nil {
  fmt.Println("Error", err)
}

// Read a data from the database
onefish := Fish{}
if err := onefishDocument.Read(&onefish); err != nil {
  fmt.Println("Error", err)
}

// Read all fish from the database, returning an array of documents.
records, err := fishCollection.ReadAll()
if err != nil {
  fmt.Println("Error", err)
}

fishies := []Fish{}
for _, f := range records {
  fishFound := Fish{}
  if err := f.Read(&onefish); err != nil {
    fmt.Println("Error", err)
  }
  fishies = append(fishies, fishFound)
}

// Delete a fish from the database
if err := onefishDocument.Delete(); err != nil {
  fmt.Println("Error", err)
}

// Delete all fish from the database
if err := fishCollection.Delete(); err != nil {
  fmt.Println("Error", err)
}

// Make a subcollection in a document
fishBabiesCollection := onefishDocument.Collection("babies")

// Make a make a document in a collection
firstbabyDocument := Document("firstbaby")

```

It is also possible to store a subcollection and data in the same document:

```go
skrillex := db.Collection("artists").Document("skrillex")

skrillex.Write(map[string]bool{
  "IsMusicGood": true,
})

skrillex.Collection("songs").Document("bangarang").Write(map[string]string{
  "Movie": "Deadpool 2",
})
```


## Documentation
- Complete documentation is available on [godoc](http://godoc.org/github.com/creativeguy2013/scribble).
- Coverage Report is available on [gocover](https://gocover.io/github.com/creativeguy2013/scribble)

## Todo/Doing
- Support for windows
- Better support for concurrency
- More methods to allow different types of reads/writes
- More tests (you can never have enough!)
