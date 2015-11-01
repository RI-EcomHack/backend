package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var customerInRangeArray []customerInRange
var currentSku string

type btnPushed struct {
	ID string
}

type customerInRange struct {
	CustomerID string
	ButtonID   string
	Date       time.Time
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/btn_pushed", buttonPushedEvent)
	router.HandleFunc("/customer_in_range/{customerId}/{beaconId}", customerInRangeEvent)
	router.HandleFunc("/flash_sale", flashSales)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func buttonPushedEvent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	decoder := json.NewDecoder(r.Body)
	var event btnPushed
	err := decoder.Decode(&event)
	if err != nil {
		panic(err)
	}

	duration, _ := time.ParseDuration("50s")

	customerIds := map[string]string{}

	for _, v := range customerInRangeArray {
		if v.ButtonID == event.ID {
			buttonTime := time.Now()
			customerTime := v.Date

			if customerTime.Add(duration).After(buttonTime) {
				customerIds[v.CustomerID] = v.CustomerID
			}
		}
	}

	for _, id := range customerIds {
		fmt.Printf("Customer %s gets a voucher for %s\n", id, currentSku)
		createCartDiscount(id, currentSku)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"success\": true}"))
	log.Printf("%s\t%s\t%s", time.Since(start), r.Method, r.RequestURI)
}

func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func customerInRangeEvent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vars := mux.Vars(r)

	event := customerInRange{
		CustomerID: vars["customerId"],
		ButtonID:   vars["beaconId"],
		Date:       time.Now(),
	}

	event.Date = time.Now()
	customerInRangeArray = append(customerInRangeArray, event)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"success\": true}"))
	log.Printf("%s\t%s\t%s", time.Since(start), r.Method, r.RequestURI)
}

// Credentials object
type Credentials struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	ProjectKey   string `json:"project_key"`
}

// Auth object
type Auth struct {
	AccessToken string `json:"access_token"`
}

func parseJson(path string) *Credentials {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		log.Fatal("Error while reading config file: ", e)
	}
	var c Credentials
	json.Unmarshal(file, &c)
	return &c
}

func parseCredentialsJson(path string) *Credentials {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		log.Fatal("Error while reading config file: ", e)
	}
	var c Credentials
	json.Unmarshal(file, &c)
	return &c
}

type sku struct {
	Results []struct {
		//		MasterVariant struct {
		//			Sku string `json:"sku"`
		//		} `json:"masterVariant"`
	} `json:"results"`
}

func flashSales(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// read credentials, shouldn't do that here
	c := parseJson("./config.json")

	// get an access_token
	auth := getAccessToken(c)
	apiUrl := fmt.Sprintf("https://api.sphere.io/%v/product-projections?limit=1&offset=%d", c.ProjectKey, 1+rand.Intn(4))

	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiUrl, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", auth.AccessToken))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error when fetching products: ", err)
	}
	defer resp.Body.Close()

	// we now use a json parsing library to decode arbitrary data
	// in order to avoid declaring structs for the entire products response
	//var dataSku sku
	//json.NewDecoder(resp.Body).Decode(&dataSku)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	var myJson ProductProjection
	json.Unmarshal(buf.Bytes(), &myJson)

	currentSku = myJson.Results[0].MasterVariant.Sku

	w.Header().Add("Access-Control-Allow-Origin", "*")

	b, _ := json.Marshal(myJson)
	// Convert bytes to string.
	w.Write(b)

	log.Printf("%s\t%s\t%s", time.Since(start), r.Method, r.RequestURI)
}

func createCartDiscount(customerId string, sku string) {
	// read credentials, shouldn't do that here
	c := parseJson("./config.json")

	// get an access_token
	auth := getAccessToken(c)
	apiUrl := fmt.Sprintf("https://api.sphere.io/%v/cart-discounts", c.ProjectKey)

	client := &http.Client{}

	json := `{
	    "name": {
	        "en": "40 percent discount"
	    },
	    "description": {
	        "en": "Flash sales 40 percent discount"
	    },
	    "value": {
	        "type": "relative",
	        "permyriad": 4000
	    },
	    "cartPredicate": "customer.id = \"%s\"",
	    "target": {
	        "type": "lineItems",
	        "predicate": "sku = \"%s\""
	    },
	    "validUntil": "2015-11-01T23:59:59+0100",
	    "sortOrder": "%s",
	    "isActive": true
	}`

	reader := strings.NewReader(fmt.Sprintf(json, customerId, sku, strconv.FormatFloat(rand.Float64(), 'f', 40, 64)+"1"))

	req, _ := http.NewRequest("POST", apiUrl, reader)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", auth.AccessToken))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error when fetching products: ", err)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("response: %s\n", string(contents))

	defer resp.Body.Close()
}

func getAccessToken(c *Credentials) *Auth {
	authUrl := fmt.Sprintf("https://%v:%v@auth.sphere.io/oauth/token", c.ClientId, c.ClientSecret)
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", fmt.Sprintf("manage_project:%v", c.ProjectKey))

	client := &http.Client{}
	req, _ := http.NewRequest("POST", authUrl, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error when retrieving access_token: ", err)
	}
	defer resp.Body.Close()
	var a Auth

	json.NewDecoder(resp.Body).Decode(&a)
	return &a
}

type ProductProjection struct {
	Count   int `json:"count"`
	Offset  int `json:"offset"`
	Results []struct {
		Categories []struct {
			ID     string `json:"id"`
			TypeID string `json:"typeId"`
		} `json:"categories"`
		CategoryOrderHints struct{} `json:"categoryOrderHints"`
		CreatedAt          string   `json:"createdAt"`
		Description        struct {
			En string `json:"en"`
		} `json:"description"`
		HasStagedChanges bool   `json:"hasStagedChanges"`
		ID               string `json:"id"`
		LastModifiedAt   string `json:"lastModifiedAt"`
		MasterVariant    struct {
			Attributes []interface{} `json:"attributes"`
			ID         int           `json:"id"`
			Images     []struct {
				Dimensions struct {
					H int `json:"h"`
					W int `json:"w"`
				} `json:"dimensions"`
				URL string `json:"url"`
			} `json:"images"`
			Prices []struct {
				ID    string `json:"id"`
				Value struct {
					CentAmount   int    `json:"centAmount"`
					CurrencyCode string `json:"currencyCode"`
				} `json:"value"`
			} `json:"prices"`
			Sku string `json:"sku"`
		} `json:"masterVariant"`
		Name struct {
			En string `json:"en"`
		} `json:"name"`
		ProductType struct {
			ID     string `json:"id"`
			TypeID string `json:"typeId"`
		} `json:"productType"`
		Published      bool     `json:"published"`
		SearchKeywords struct{} `json:"searchKeywords"`
		Slug           struct {
			En string `json:"en"`
		} `json:"slug"`
		TaxCategory struct {
			ID     string `json:"id"`
			TypeID string `json:"typeId"`
		} `json:"taxCategory"`
		Variants []interface{} `json:"variants"`
		Version  int           `json:"version"`
	} `json:"results"`
	Total int `json:"total"`
}
