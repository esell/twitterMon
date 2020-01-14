package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"git.sr.ht/~tslocum/cview"
	"github.com/ChimeraCoder/anaconda"
	"github.com/gdamore/tcell"
)

func createTextViewItem(title string) *cview.TextView {
	tempView := cview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false).
		SetWordWrap(true)
	tempView.SetBorder(true).SetTitle(title)
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
	for k, v := range columns {
		if view == v {
			if k+1 == len(columns) {
				return columns[0]
			} else {
				return columns[k+1]
			}
		}
	}
	return nil
}
