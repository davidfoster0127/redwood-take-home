package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var (
	batteryPickups = map[string]*BatteryPickup{} // stored battery pickups
)

type BatteryPickup struct {
	ID        uuid.UUID `json:"id"`
	Location  string    `json:"location"`
	Batteries []Battery `json:"batteries"`
	Quote     float32   `json:"quote"`
	Accepted  bool      `json:"accepted"`
}

// {
// 	"id": "xxxx",
// 	"location": "here",
// 	"batteries": [
// 		{
// 			"type": 1,
// 			"condition": 1,
// 			"weight": 30,
// 			"capacity": 50,
// 		},
// 		{
// 			"type": 2,
// 			"condition": 3,
// 			"weight": 10,
// 			"capacity": 20,
// 		}
// 	],
// 	"quote": 0,
// 	"accepted": false
// }

type Battery struct {
	Type      int     `json:"type"`
	Condition int     `json:"condition"`
	Weight    float32 `json:"weight"`
	Capacity  float32 `json:"capacity"`
}

// {
// 	"type": 1,
// 	"condition": 1,
// 	"weight": 30,
// 	"capacity": 50
// }

func newPickup(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
	switch r.Method {
	case http.MethodPost: // create new battery pickup
		var newPickup = &BatteryPickup{}
		var decoder = json.NewDecoder(r.Body)
		var err = decoder.Decode(newPickup)
		if err != nil {
			log.Printf("err = %s", err)
			http.Error(w, "Unable to decode body", http.StatusBadRequest)
			return
		}

		// verify required pickup request fields
		if newPickup.Location == "" { // loosely verified location...
			http.Error(w, "location required", http.StatusBadRequest)
			return
		}

		if len(newPickup.Batteries) == 0 {
			http.Error(w, "list of batteries required", http.StatusBadRequest)
			return
		}

		newPickup.ID = uuid.New() // assign new id to pickup

		// verify batteries in pickup request, price them, and create a quote
		var newQuote float32 = 0
		for _, battery := range newPickup.Batteries {
			if battery.Type == 0 {
				http.Error(w, "invalid battery type", http.StatusBadRequest)
				return
			}

			if battery.Weight == 0 {
				http.Error(w, "invalid battery weight", http.StatusBadRequest)
				return
			}

			if battery.Capacity == 0 {
				http.Error(w, "invalid battery capacity", http.StatusBadRequest)
				return
			}

			if battery.Condition == 0 {
				http.Error(w, "invalid battery condition", http.StatusBadRequest)
				return
			}

			batteryPrice := currentRules.priceBattery(battery)
			newQuote += batteryPrice
		}

		newPickup.Quote = newQuote

		batteryPickups[newPickup.ID.String()] = newPickup // store pickup request in memory

		w.WriteHeader(http.StatusCreated)
		var encoder = json.NewEncoder(w)
		err = encoder.Encode(newPickup)
		if err != nil {
			log.Printf("err = %s", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func pickups(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
	id := strings.TrimPrefix(r.URL.Path, "/pickup/") // extract pickup id

	switch r.Method {
	case http.MethodGet: // return information on existing battery pickup(s)
		if id == "" { // return all pickup requests
			pickups := struct {
				Pickups []BatteryPickup
			}{}

			acceptedOnly := r.FormValue("accept") // optional parameter to return only pickups that are accepted
			for _, pickup := range batteryPickups {
				if acceptedOnly == "true" || acceptedOnly == "t" || acceptedOnly == "yes" {
					if pickup.Accepted {
						pickups.Pickups = append(pickups.Pickups, *pickup)
					}
				} else {
					pickups.Pickups = append(pickups.Pickups, *pickup)
				}

			}
			var encoder = json.NewEncoder(w)
			var err = encoder.Encode(pickups)
			if err != nil {
				log.Printf("err = %s", err)
				http.Error(w, "error encoding response", http.StatusInternalServerError)
				return
			}
			return
		}

		// return specific pickup request
		pickup, exists := batteryPickups[id]
		if !exists {
			http.Error(w, "pickup not found", http.StatusNoContent)
			return
		}

		var encoder = json.NewEncoder(w)
		var err = encoder.Encode(pickup)
		if err != nil {
			log.Printf("err = %s", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
		return
	case http.MethodPatch: // update existing pickup information i.e. to accept or reject quote
		if len(id) == 0 {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		pickup, exists := batteryPickups[id]
		if !exists {
			http.Error(w, "pickup not found", http.StatusNoContent)
			return
		}

		// verify pickup request patch has required field "accept"
		accept := r.FormValue("accept")
		if accept == "" {
			http.Error(w, "missing accept parameter", http.StatusBadRequest)
			return
		}

		if accept == "true" || accept == "t" || accept == "yes" {
			pickup.Accepted = true
			io.WriteString(w, "pickup accepted\n")
		} else if accept == "false" || accept == "f" || accept == "no" { // allow rejection of previously accepted pickup request
			pickup.Accepted = false
			io.WriteString(w, "pickup rejected\n")
		} else {
			http.Error(w, "invalid accept parameter", http.StatusBadRequest)
			return
		}
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
