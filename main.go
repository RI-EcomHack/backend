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

type btnPushed struct {
	ID   string
	Date string
	Time string
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/btn_pushed", buttonPushed)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func stringsToSet(strings []string) mapset.Set {
	set := mapset.NewThreadUnsafeSet()
	for _, s := range strings {
		set.Add(s)
	}

	return set
}

func buttonPushed(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	decoder := json.NewDecoder(r.Body)
	var t btnPushed
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	btnPushedArray = append(btnPushedArray, t)
	fmt.Printf("A button was pushed: %+v\n", i)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"success\": true}"))
	log.Printf("%s\t%s\t%s", time.Since(start), r.Method, r.RequestURI)
}
