package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
)

func createTextViewItem(title string) *cview.TextView {
	tempView := cview.NewTextView()
	tempView.SetDynamicColors(true)
	tempView.SetRegions(false)
	tempView.SetWordWrap(true)
	tempView.SetBorder(true)
	tempView.SetTitle(title)
	tempView.SetBorderColor(tcell.ColorGrey)
	tempView.SetChangedFunc(func() {
		tempView.ScrollToBeginning()
		app.Draw()
	})

	tempView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			blah := nextColumn(columnItems, tempView)
			app.SetFocus(blah)
			tempView.SetTitleColor(tcell.ColorWhite)
			blah.SetTitleColor(tcell.ColorYellow)
			return nil
		}
		return event
	})
	return tempView
}

func textViewProcessFeed(tv *cview.TextView, listID int64, api *anaconda.TwitterApi) {
	for {
		tv.Clear()
		lastUpdateTime := time.Now()
		fmt.Fprintf(tv, "[#CCFFCC::b]%s %v[white]\n\n", "UPDATED", lastUpdateTime.Format(time.UnixDate))
		timeTweets, err := getListTweets(api, listID)
		if err != nil {
			log.Println(err)
		}
		tweetSplit := buildBreak(tv)
		for _, tweet := range timeTweets {
			tempTime, err := tweet.CreatedAtTime()
			if err != nil {
				log.Println(err)
			}
			if tweet.RetweetedStatus != nil {
				fmt.Fprintf(tv, "[#66CCFF::b]@%s[-::-] at [#66CCFF]%s[-] [red]*RT*[-]:\n\n%s\n%s\n\n", tweet.User.ScreenName, tempTime.Local().Format(time.UnixDate), tweet.FullText, tweetSplit)
			} else {
				fmt.Fprintf(tv, "[#66CCFF::b]@%s[-::-] at [#66CCFF]%s[-]:\n\n%s\n%s\n\n", tweet.User.ScreenName, tempTime.Local().Format(time.UnixDate), tweet.Text, tweetSplit)
			}
		}
		time.Sleep(parsedconfig.Refresh * time.Minute)
	}
}

func showModal(currentColumn *cview.TextView) {
	buttons := make([]string, 0)

	for _, column := range columnItems {
		if !isColumnActive(column) {
			buttons = append(buttons, column.GetTitle())
		}
	}
	buttons = append(buttons, "Close")

	modal := cview.NewModal()
	modal.SetText("Select a column to add")
	modal.AddButtons(buttons)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Close" {
			app.SetRoot(flex, true)
		} else {
			columnToAdd := getColumnByTitle(buttonLabel)
			app.QueueUpdateDraw(func() {
				flex.AddItem(columnToAdd, 0, 1, true)
				app.SetRoot(flex, true)
				app.SetFocus(columnToAdd)
				currentColumn.SetTitleColor(tcell.ColorWhite)
				columnToAdd.SetTitleColor(tcell.ColorYellow)
			})
		}
	})
	app.QueueUpdateDraw(func() {
		app.SetRoot(modal, true)
		app.SetFocus(modal)
	})
}

func buildBreak(tv *cview.TextView) string {
	_, _, w, _ := tv.GetInnerRect()
	charCount := w / 2
	var sb strings.Builder
	for i := 0; i < charCount/2; i++ {
		sb.WriteString(" ")
	}

	for i := 0; i < charCount; i++ {
		sb.WriteString("=")
	}
	return sb.String()
}

func nextColumn(columns []*cview.TextView, view *cview.TextView) *cview.TextView {
	childers := flex.GetChildren()
	for k, v := range childers {
		if view == v.(*cview.TextView) {
			if k+1 == len(childers) {
				return childers[0].(*cview.TextView)
			} else {
				return childers[k+1].(*cview.TextView)
			}
		}
	}
	return nil
}

func getColumnByTitle(title string) *cview.TextView {
	for _, column := range columnItems {
		if column.GetTitle() == title {
			return column
		}
	}
	return nil
}

func isColumnActive(column *cview.TextView) bool {
	childers := flex.GetChildren()
	for _, v := range childers {
		if column == v.(*cview.TextView) {
			return true
		}
	}

	return false
}
