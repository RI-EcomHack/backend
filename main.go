package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/gorilla/mux"
)

var btnPushedArray []btnPushed
var customerInRangeArray []customerInRange

type btnPushed struct {
	ID   string
	Date string
	Time string
}

type customerInRange struct {
	CustomerID string
	ButtonID   string
	Date       string
	Time       string
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/btn_pushed", buttonPushedEvent)
	router.HandleFunc("/customer_in_range", customerInRangeEvent)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func stringsToSet(strings []string) mapset.Set {
	set := mapset.NewThreadUnsafeSet()
	for _, s := range strings {
		set.Add(s)
	}

	return set
}

func buttonPushedEvent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	decoder := json.NewDecoder(r.Body)
	var event btnPushed
	err := decoder.Decode(&event)
	if err != nil {
		panic(err)
	}

	btnPushedArray = append(btnPushedArray, event)
	fmt.Printf("Button %s was pushed\n", event.ID)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"success\": true}"))
	log.Printf("%s\t%s\t%s", time.Since(start), r.Method, r.RequestURI)
}

func customerInRangeEvent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	decoder := json.NewDecoder(r.Body)
	var event customerInRange
	err := decoder.Decode(&event)
	if err != nil {
		panic(err)
	}

	customerInRangeArray = append(customerInRangeArray, event)
	fmt.Printf("Customer %s is in range of button %s\n", event.CustomerID, event.ButtonID)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"success\": true}"))
	log.Printf("%s\t%s\t%s", time.Since(start), r.Method, r.RequestURI)
}
