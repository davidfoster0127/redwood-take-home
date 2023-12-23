#!/bin/bash

printf "running test.sh\n"

curl --location 'http://localhost:8080/'

printf "\npost new pricing rules to server\n"
curl --location 'http://localhost:8080/pricing_rules' \
--header 'Content-Type: text/plain' \
--data '{
	"type_map": {
		"1": 50,
		"2": 100,
		"3": 150
	},
	"condition_map": {
		"1": 1,
		"2": 0.8,
		"3": 0.5
	},
	"weight_factor": 0.6,
	"capacity_factor": 0.7
}'

printf "\nget existing pricing rules\n"
curl --location 'http://localhost:8080/pricing_rules'

printf "\npost new pickup request\n"
response=$(curl -s --location 'http://localhost:8080/pickup' \
--header 'Content-Type: text/plain' \
--data '{
	"location": "here",
	"batteries": [
		{
			"type": 1,
			"condition": 1,
			"weight": 30,
			"capacity": 50
		},
		{
			"type": 2,
			"condition": 3,
			"weight": 10,
			"capacity": 20
		}
    ]
}')
printf "$response\n"

# extracting pickup id (requires jq, please install on host if not already)
id=$(printf "$response" | jq -r '.id')

printf "\nget pickup request\n"
curl --location "http://localhost:8080/pickup/$id"

printf "\naccept quote for pickup request\n"
curl --location --request PATCH "http://localhost:8080/pickup/$id" \
--form 'accept="true"'

printf "\nget all existing pickup requests\n"
curl --location "http://localhost:8080/pickup/"

printf "\nget all existing pickup requests that have been accepted\n"
curl --location "http://localhost:8080/pickup/?accept=true"
