package main

import (
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

func getAPIClient(accessToken, accessSecret, consumerKey, consumerSecret string) *anaconda.TwitterApi {
	api := anaconda.NewTwitterApiWithCredentials(accessToken, accessSecret, consumerKey, consumerSecret)
	return api
}

func getFollowing(api *anaconda.TwitterApi) ([]anaconda.User, error) {
	following := make([]anaconda.User, 0)
	urlVals := url.Values{}
	urlVals.Set("count", "200")
	pages := api.GetFriendsListAll(urlVals)
	for page := range pages {
		following = append(following, page.Friends...)
	}

	return following, nil
}

func createList(api *anaconda.TwitterApi, listName string) (anaconda.List, error) {
	urlVals := url.Values{}
	urlVals.Set("count", "200")
	urlVals.Set("mode", "private")
	listResult, err := api.CreateList(listName, "", urlVals)
	if err != nil {
		return anaconda.List{}, err
	}
	return listResult, nil
}

func getListTweets(api *anaconda.TwitterApi, listID int64) ([]anaconda.Tweet, error) {
	urlVals := url.Values{}
	urlVals.Set("count", "50")
	tweets, err := api.GetListTweets(listID, true, urlVals)
	if err != nil {
		return make([]anaconda.Tweet, 0), err
	}

	return tweets, nil
}
