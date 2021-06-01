package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
)

const (
	apiBase = "https://hacker-news.firebaseio.com/v0"
)

var templates = template.Must(template.ParseFiles("./views/index.gohtml"))

type HNItem struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Url       string `json:"url"`
	PostType  string `json:"type"`
	SortOrder int
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

	//sort the stories according to order
	sort.Slice(HNItems[:], func(i, j int) bool {
		return HNItems[i].SortOrder < HNItems[j].SortOrder
	})

	HNItems = HNItems[:30]

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

	//use go routine and channel here to get the top 30 stories
	var wg sync.WaitGroup
	ch := make(chan HNItem)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go GetAndStoreData(ch, ids[i], i, &wg)
	}

	//wrapping the waitgroup and close channel in a goroutine is important
	go func() {
		wg.Wait()
		close(ch)
	}()

	for {
		res, ok := <-ch
		if ok == false {
			break
		}
		HNItems = append(HNItems, res)
	}

	return HNItems
}

func GetAndStoreData(ch chan HNItem, id int, order int, wg *sync.WaitGroup) {
	defer wg.Done()
	item, err := GetItemData(id)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if item.PostType == "story" && len(item.Url) > 0 {
			item.SortOrder = order
			ch <- item
		}
	}
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
