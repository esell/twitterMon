package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/mattn/go-sqlite3"
)

func insertFollowing(acctName, list string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into following(acct_name, assigned_list) values(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(acctName, list)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func insertList(list, slug string, id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into lists(list_name, slug, id) values(?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(list, slug, id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func loadFollowing(fileName string) error {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	r := csv.NewReader(strings.NewReader(string(file)))
	r.FieldsPerRecord = -1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(record) == 2 {
			err = insertFollowing(record[1], "")
			if err != nil {
				log.Printf("error inserting %s: %v\n", record[1], err)
			}
		} else if len(record) == 3 {
			err = insertFollowing(record[1], record[2])
			if err != nil {
				log.Printf("error inserting %s: %v\n", record[1], err)
			}
		}
	}

	return nil
}

func loadCreateLists(api *anaconda.TwitterApi, fileName string) error {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	r := csv.NewReader(strings.NewReader(string(file)))
	r.FieldsPerRecord = -1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		newList, err := createList(api, record[0])
		if err != nil {
			log.Println(err)
		}

		err = insertList(newList.Name, newList.Slug, newList.Id)
		if err != nil {
			log.Printf("error inserting %s: %v\n", record[0], err)
		}
	}

	return nil
}

func assignToList(api *anaconda.TwitterApi, listName string, listID int64) {
	acctsToAdd := make([]string, 0)
	rows, err := db.Query("select acct_name from following where assigned_list = '" + listName + "'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// map of list names -> list ids
	for rows.Next() {
		var acctName string
		rows.Scan(&acctName)
		acctsToAdd = append(acctsToAdd, acctName)
	}
	api.AddMultipleUsersToList(acctsToAdd, listID, nil)
}

func assignAllToLists(api *anaconda.TwitterApi) {

	listIDs := make(map[string]int64)
	rows, err := db.Query("select list_name,id from lists")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// map of list names -> list ids
	for rows.Next() {
		var listName string
		var listID int64
		rows.Scan(&listName, &listID)
		listIDs[listName] = listID
	}

	for k, v := range listIDs {
		assignToList(api, k, v)
	}
}
func dumpLists() {
	rows, err := db.Query("select list_name,id from lists")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// map of list names -> list ids
	for rows.Next() {
		var listName string
		var listID int64
		rows.Scan(&listName, &listID)
		fmt.Printf("%s,%d\n", listName, listID)
	}
}
