/*
REFERENCES:
1) https://github.com/HackerNews/API1
*/

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	apiBase = "https://hacker-news.firebaseio.com/v0"
)

var templates = template.Must(template.ParseFiles("./views/index.gohtml"))

type HNItem struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Url      string `json:"url"`
	PostType string `json:"type"`
	// Author      string    `json:"by"`
	// TimeCreated time.Time `json:"time"`
}

func main() {
	Start()
}

func Start() {
	http.HandleFunc("/", TopStoriesHandler)
	fmt.Println("Starting the website.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func TopStoriesHandler(w http.ResponseWriter, r *http.Request) {
	var HNItems []HNItem

	HNItems = GetTopStories()

	data := struct {
		HNItems []HNItem
	}{
		HNItems: HNItems,
	}

	err := templates.ExecuteTemplate(w, "index.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetTopStories() []HNItem {
	var ids []int
	var HNItems []HNItem

	//top stories api url:
	topStoriesApi := apiBase + "/topstories.json"

	resp, err := http.Get(topStoriesApi)
	if err != nil {
		fmt.Println("An error occured when calling the API.")
	}

	defer resp.Body.Close()

	//returns an array of 500 ids
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(bodyBytes, &ids)

	//get the top 30 stories
	i := 0
	for len(HNItems) < 30 {
		itemData, _ := GetItemData(ids[i])
		if itemData.PostType == "story" {
			HNItems = append(HNItems, itemData)
		}
		i++
	}

	return HNItems
}

func GetItemData(id int) (HNItem, error) {
	var item HNItem
	itemApi := fmt.Sprintf("%s/item/%d.json", apiBase, id)
	resp, err := http.Get(itemApi)
	if err != nil {
		return item, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&item)
	if err != nil {
		return item, err
	}
	return item, nil
}
