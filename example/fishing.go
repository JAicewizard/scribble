package main

import (
	"fmt"

	scribble "github.com/CreativeGuy2013/scribble"
)

//Fish is a fish struct
type Fish struct{ Name string }

func main() {

	dir := "./db"

	db, err := scribble.New(dir)
	if err != nil {
		fmt.Println("Error 1", err)
	}

	fishCollection := db.Collection("fish")

	// Write a fish to the database
	for _, name := range []string{"onefish", "twofish", "redfish", "bluefish"} {
		if err := fishCollection.Document(name).Write(Fish{Name: name}); err != nil {
			fmt.Println("Error 2", err)
		}
	}

	fishCollection.Document("purple").Collection("teeth").Document("3").Write(map[string]string{
		"one": "two",
		"two": "three",
	})

	// Read a fish from the database (passing fish by reference)
	onefish := Fish{}
	if err := fishCollection.Document("onefish").Read(&onefish); err != nil {
		fmt.Println("Error 3", err)
	}

	// Read all fish from the database, unmarshaling the response.
	records, err := fishCollection.GetDocuments()
	if err != nil {
		fmt.Println("Error 4", err)
	}

	fishies := []Fish{}
	for _, f := range records {
		fishFound := Fish{}
		if err := f.Read(&fishFound); err != nil {
			fmt.Println("Error 5", err)
		}
		fishies = append(fishies, fishFound)
	}

	// // Delete a fish from the database
	// if err := db.Delete("fish", "onefish"); err != nil {
	// 	fmt.Println("Error", err)
	// }
	//
	// // Delete all fish from the database
	// if err := db.Delete("fish", ""); err != nil {
	// 	fmt.Println("Error", err)
	// }

}
