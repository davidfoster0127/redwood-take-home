package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var (
	currentRules = PricingRules{}
)

type PricingRules struct {
	TypeMap        map[int]float32 `json:"type_map"`
	ConditionMap   map[int]float32 `json:"condition_map"`
	WeightFactor   float32         `json:"weight_factor"`
	CapacityFactor float32         `json:"capacity_factor"`
}

// {
// 	"type_map": {
// 		"1": 50,
// 		"2": 100,
// 		"3": 150
// 	},
// 	"condition_map": {
// 		"1": 1,
// 		"2": 0.8,
// 		"3": 0.5,
// 	},
// 	"weight_factor": 0.6,
// 	"capacity_factor": 0.7
// }

func (rules PricingRules) priceBattery(battery Battery) float32 {
	price := rules.TypeMap[battery.Type]
	price = price * rules.ConditionMap[battery.Condition]
	price = price + (rules.WeightFactor * battery.Weight)
	price = price + (rules.CapacityFactor * battery.Capacity)
	return price
}

func pricingRules(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
	switch r.Method {
	case "POST": // ingest new rules
		var newRules = &PricingRules{}
		var decoder = json.NewDecoder(r.Body)
		var err = decoder.Decode(newRules)
		if err != nil {
			log.Printf("err = %s", err)
			http.Error(w, "Unable to decode body", http.StatusBadRequest)
			return
		}

		// check for valid pricing rules document
		if len(newRules.TypeMap) == 0 {
			http.Error(w, "TypeMap not found", http.StatusBadRequest)
			return
		}

		if len(newRules.ConditionMap) == 0 {
			http.Error(w, "ConditionMap not found", http.StatusBadRequest)
			return
		}

		if newRules.CapacityFactor == 0 {
			http.Error(w, "CapacityFactor not found", http.StatusBadRequest)
			return
		}

		if newRules.WeightFactor == 0 {
			http.Error(w, "WeightFactor not found", http.StatusBadRequest)
			return
		}

		// set currentRules to new rules
		currentRules = *newRules

		w.WriteHeader(http.StatusCreated)
		var encoder = json.NewEncoder(w)
		err = encoder.Encode(currentRules)
		if err != nil {
			log.Printf("err = %s", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
		return
	case "GET": // return existing rules
		var encoder = json.NewEncoder(w)
		var err = encoder.Encode(currentRules)
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
