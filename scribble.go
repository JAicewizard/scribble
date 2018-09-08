// Package scribble is a tiny JSON database
package scribble

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Version is the current version of the project
const Version = "2.1.0"

type (

	// Logger is a generic logger interface
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	//Collection a collection of documents
	Collection struct {
		dir string // the directory where scribble will create the database
		err error
	}

	//Document a single document which can have sub collections
	Document struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		err     error
	}
)

// Options uses for specification of working golang-scribble
type Options struct {
	Logger // the logger scribble will use (configurable)
}

// New creates a new scribble database at the desired directory location, and
// returns a *Driver to then use for interacting with the database
func New(dir string) (*Document, error) {
	//Clean the filepath before using it
	dir = filepath.Clean(dir)

	document := Document{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
	}

	// if the collection doesn't exist create it
	if _, err := os.Stat(filepath.Join(document.dir, "doc.json")); err == nil {
		return &document, nil
	}

	if _, err := os.Stat(document.dir); err != nil {
		if err := os.MkdirAll(document.dir, 0755); err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	}

	// if the document doesn't exist create it
	return &document, ioutil.WriteFile(filepath.Join(document.dir, "doc.json"), []byte("{}"), 0644)
}

//Document gets a document from a collection
func (c *Collection) Document(key string) *Document {
	if key == "" {
		return &Document{
			dir:     c.dir,
			mutexes: make(map[string]*sync.Mutex),
			err:     fmt.Errorf("key for document is empty"),
		}
	} else if c.err != nil {
		return &Document{
			dir:     c.dir,
			mutexes: make(map[string]*sync.Mutex),
			err:     c.err,
		}
	}

	dir := filepath.Join(c.dir, key)

	document := Document{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
	}

	return &document
}

//Collection gets a collction from in a document
func (d *Document) Collection(name string) *Collection {
	if name == "" {
		return &Collection{
			dir: d.dir,
			err: fmt.Errorf("name for collection is empty"),
		}
	} else if d.err != nil {
		return &Collection{
			dir: d.dir,
			err: d.err,
		}
	}

	dir := filepath.Join(d.dir, name)

	collection := Collection{
		dir: dir,
	}

	return &collection
}

// Write locks the database and attempts to write the record to the database under
// the [collection] specified with the [resource] name given
func (d *Document) Write(v interface{}) error {
	// check if there was an error
	if err := d.Check(); err != nil {
		return err
	}

	// ensure there is a place to save record
	if d.dir == "" {
		return fmt.Errorf("missing document - no place to save record")
	}

	if _, err := os.Stat(d.dir); err != nil {
		if err := os.MkdirAll(d.dir, 0755); err != nil {
			return err
		}
	}

	mutex := d.getOrCreateMutex()
	mutex.Lock()
	defer mutex.Unlock()

	//
	dir := d.dir
	fnlPath := filepath.Join(dir, "doc.json")
	tmpPath := fnlPath + ".tmp"

	// create collection directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	//
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	// write marshaled data to the temp file
	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	// move final file into place
	return os.Rename(tmpPath, fnlPath)
}

// Read a record from the database
func (d *Document) Read(v interface{}) error {
	// check if there was an error
	if err := d.Check(); err != nil {
		return err
	}

	// ensure there is a place to save record
	if d.dir == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	//
	record := filepath.Join(d.dir, "doc.json")

	// check to see if file exists
	if _, err := stat(record); err != nil {
		return err
	}

	// read record from database
	b, err := ioutil.ReadFile(record)
	if err != nil {
		return err
	}

	// unmarshal data
	return json.Unmarshal(b, &v)
}

// GetDocuments gets all documents in a collection.
func (c *Collection) GetDocuments() ([]*Document, error) {
	// check if there was an error
	if err := c.Check(); err != nil {
		return nil, err
	}

	// ensure there is a collection to read
	if c.dir == "" {
		return nil, fmt.Errorf("missing collection - unable to record location")
	}

	//
	dir := c.dir

	// check to see if collection (directory) exists
	if _, err := stat(dir); err != nil {
		return nil, err
	}

	// read all the files in the transaction.Collection; an error here just means
	// the collection is either empty or doesn't exist
	files, _ := ioutil.ReadDir(dir)

	// the files read from the database
	var records []*Document

	// iterate over each of the files, and add the resulting document to records
	for _, file := range files {
		// append read file
		records = append(records, &Document{
			dir:     filepath.Join(dir, file.Name()),
			mutexes: make(map[string]*sync.Mutex),
		})
	}

	// unmarhsal the read files as a comma delimeted byte array
	return records, nil
}

// Delete locks that database and removes the document including all of its sub documents
func (d *Document) Delete() error {
	// check if there was an error
	if err := d.Check(); err != nil {
		return err
	}

	//
	mutex := d.getOrCreateMutex()
	mutex.Lock()
	defer mutex.Unlock()

	//
	dir := d.dir

	switch fi, err := stat(dir); {

	// if fi is nil or error is not nil return
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file or directory named %v", dir)

	// remove directory and all contents
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)

	// remove file
	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}

	return nil
}

// Delete removes a collection and all of its childeren
func (c *Collection) Delete() error {
	// check if there was an error
	if err := c.Check(); err != nil {
		return err
	}

	//
	dir := c.dir

	switch fi, err := stat(dir); {

	// if fi is nil or error is not nil return
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file or directory named %v", dir)

	// remove directory and all contents
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)

	// remove file
	case fi.Mode().IsRegular():
		return os.RemoveAll(filepath.Join(dir, "doc.json"))
	}

	return nil
}

//Check if there is an error while getting the collection
func (c *Collection) Check() error {
	return c.err
}

//Check if there is an error while getting the document
func (d *Document) Check() error {
	return d.err
}

//
func stat(path string) (fi os.FileInfo, err error) {

	// check for dir, if path isn't a directory check to see if it's a file
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(filepath.Join(path, "doc.json"))
	}

	return
}

// getOrCreateMutex creates a new collection specific mutex any time a collection
// is being modfied to avoid unsafe operations
func (d *Document) getOrCreateMutex() *sync.Mutex {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	m, ok := d.mutexes[d.dir]

	// if the mutex doesn't exist make it
	if !ok {
		m = &sync.Mutex{}
		d.mutexes[d.dir] = m
	}

	return m
}
