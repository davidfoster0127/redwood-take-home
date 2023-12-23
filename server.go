package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", root)
	http.HandleFunc("/pickup", newPickup)           // create new pickup
	http.HandleFunc("/pickup/", pickups)            // get or patch existing pickups
	http.HandleFunc("/pricing_rules", pricingRules) // get current pricing rules or set new ones

	log.Println("running server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func root(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
	io.WriteString(w, "server running\n")
}
