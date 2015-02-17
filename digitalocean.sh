#!/bin/bash
#update an A record in digital Ocean. Dynamic DNS style.
#API info here:
#https://developers.digitalocean.com/#domains-list
#https://gist.github.com/glpunk/8679208#comment-1274395

#your domain ID
domain_id="$2"
#base domain
domain="`echo "$3" | rev | cut -f-2 -d'.' | rev`"
#your api key
api_key="$1"
#new value
host="$4"

### don't change ###
echo content="$(curl -H "Authorization: Bearer $api_key" -H "Content-Type: application/json" \
                 -d '{"data": "'"$host"'"}' \
                 -X PUT "https://api.digitalocean.com/v2/domains/$domain/records/$domain_id")"
