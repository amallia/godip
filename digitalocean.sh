#!/bin/bash
#update an A record in digital Ocean. Dynamic DNS style.
#API info here:
#https://developers.digitalocean.com/#domains-list
#https://gist.github.com/glpunk/8679208#comment-1274395

#your domain ID
domain_id= $1 | cut -f2- -d'.'
#record to update
record_id=$1 | cut -f1 -d'.'
#your api key
api_key= $0

### don't change ###
echo content="$(curl -H "Authorization: Bearer $api_key" -H "Content-Type: application/json" \
                 -d '{"type": "A", "name": "'"$record_id"'", "data": "'"$2"'"}' \
                 -X POST "https://api.digitalocean.com/v2/domains/$domain_id/records")"
