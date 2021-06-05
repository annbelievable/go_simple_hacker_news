package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const (
	apiBase = "https://hacker-news.firebaseio.com/v0"
)

var templates = template.Must(template.ParseFiles("./views/index.gohtml"))
var goCache = cache.New(5*time.Minute, 10*time.Minute)

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

	//if there is any error occured, display a different template
	HNItems, err := GetTopStories()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(HNItems) > 1 {
		//sort the stories according to order
		sort.Slice(HNItems[:], func(i, j int) bool {
			return HNItems[i].SortOrder < HNItems[j].SortOrder
		})
		//get only 30 items
		HNItems = HNItems[:30]
	}

	data := struct {
		HNItems []HNItem
	}{
		HNItems: HNItems,
	}

	err = templates.ExecuteTemplate(w, "index.gohtml", data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//update the func to return error too
func GetTopStories() ([]HNItem, error) {
	var ids []int
	var HNItems []HNItem

	//top stories api url:
	topStoriesApi := apiBase + "/topstories.json"

	//returns an array of 500 ids
	resp, err := http.Get(topStoriesApi)
	if err != nil {
		return HNItems, err
	}

	if resp.StatusCode != 200 {
		return HNItems, errors.New("Issue with calling top stories API.")
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return HNItems, err
	}

	err = json.Unmarshal(bodyBytes, &ids)
	if err != nil {
		return HNItems, err
	}

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

	return HNItems, nil
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
	idString := strconv.Itoa(id)

	if val, found := goCache.Get(idString); found {
		itemPointer := val.(*HNItem)
		return *itemPointer, nil
	}

	itemApi := fmt.Sprintf("%s/item/%d.json", apiBase, id)
	resp, err := http.Get(itemApi)
	defer resp.Body.Close()
	if err != nil {
		return item, err
	}

	if resp.StatusCode != 200 {
		return item, errors.New("Issue with calling item API.")
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&item)
	if err != nil {
		return item, err
	}

	goCache.Set(idString, &item, cache.DefaultExpiration)

	return item, nil
}
