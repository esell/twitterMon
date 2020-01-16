package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"git.sr.ht/~tslocum/cview"
	"github.com/ChimeraCoder/anaconda"
	"github.com/gdamore/tcell"
	_ "github.com/mattn/go-sqlite3"
)

type conf struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
	Refresh        time.Duration
	Lists          []list
}

type list struct {
	Name   string
	ListID int64
}

var (
	configFile     = flag.String("c", "conf.json", "config file location")
	createDb       = flag.Bool("n", false, "create new, empty database")
	createLists    = flag.Bool("l", false, "create empty twitter lists")
	doGetFollowing = flag.Bool("f", false, "dump csv list of who you're following")
	doDumpLists    = flag.Bool("d", false, "dump lists from db")
	db             *sql.DB
	parsedconfig   = conf{}
	columnItems    = make([]*cview.TextView, 0)
	app            = cview.NewApplication()
	flex           = cview.NewFlex()
)

func main() {

	flag.Parse()
	readConfig()

	// twitter API client
	apiClient := getAPIClient(parsedconfig.AccessToken, parsedconfig.AccessSecret, parsedconfig.ConsumerKey, parsedconfig.ConsumerSecret)
	apiClient.SetLogger(anaconda.BasicLogger)
	second := 2
	duration := time.Duration(second) * time.Second
	apiClient.EnableThrottling(duration, 1)

	// no db access needed for this...
	if *doGetFollowing {
		following, err := getFollowing(apiClient)
		if err != nil {
			log.Fatal(err)
		}

		for _, friend := range following {
			fmt.Printf("\"%s\",\"@%s\"\n", friend.Name, friend.ScreenName)
		}
		os.Exit(0)
	}

	if *createDb {
		var err error
		log.Println("creating new database...")
		os.Remove("./tm.db")

		db, err = sql.Open("sqlite3", "./tm.db")
		if err != nil {
			log.Fatal(err)
		}

		sqlStmt := `
                          create table following (acct_name text not null primary key, assigned_list text);
                          delete from following;
                          create table lists (list_name text not null primary key, slug text, id integer);
                          delete from lists;
                          `
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlStmt)
		}

		// import data
		err = loadFollowing("out.csv")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		var err error
		db, err = sql.Open("sqlite3", "./tm.db")
		if err != nil {
			log.Fatal(err)
		}
	}
	defer db.Close()

	// create empty lists in Twitter
	if *createLists {
		err := loadCreateLists(apiClient, "lists.csv")
		if err != nil {
			log.Fatal(err)
		}
		assignAllToLists(apiClient)
		os.Exit(0)
	}

	if *doDumpLists {
		dumpLists()
		os.Exit(0)
	}
	// start up console view
	cview.Borders.HorizontalFocus = cview.BoxDrawingsLightHorizontal
	cview.Borders.VerticalFocus = cview.BoxDrawingsLightVertical
	cview.Borders.TopLeftFocus = cview.BoxDrawingsLightDownAndRight
	cview.Borders.TopRightFocus = cview.BoxDrawingsLightDownAndLeft
	cview.Borders.BottomLeftFocus = cview.BoxDrawingsLightUpAndRight
	cview.Borders.BottomRightFocus = cview.BoxDrawingsLightUpAndLeft

	// key commands
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlA {
			currentColumn := app.GetFocus()
			showModal(currentColumn.(*cview.TextView))
		}
		if event.Key() == tcell.KeyCtrlD {
			inFocus := app.GetFocus()
			target := nextColumn(columnItems, inFocus.(*cview.TextView))
			flex.RemoveItem(inFocus)
			app.SetFocus(target)
			target.SetTitleColor(tcell.ColorYellow)
		}
		return event
	})

	for _, list := range parsedconfig.Lists {
		tempView := createTextViewItem(list.Name)
		columnItems = append(columnItems, tempView)
		go textViewProcessFeed(tempView, list.ListID, apiClient)
	}

	for k, v := range columnItems {
		if k == 0 {
			v.SetTitleColor(tcell.ColorYellow)
		}
		// hack
		flex.AddItem(v, 0, 1, true)
	}

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal(err)
	}
}

func readConfig() {
	file, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("unable to read config file, exiting...")
	}
	if err := json.Unmarshal(file, &parsedconfig); err != nil {
		log.Fatal("unable to marshal config file, exiting...")
	}
}
