
#!/bin/bash 

KEY=$(<~/.grafanakey)
HOST="http://localhost:3000"

for dash in $(curl -sSL -k -H "Authorization: Bearer $KEY" $HOST/api/search\?query\=\& | jq '.' |grep -i uri|awk -F '"uri": "' '{ print $2 }'|awk -F '"' '{print $1 }'); do
  curl -sSL -k -H "Authorization: Bearer ${KEY}" "${HOST}/api/dashboards/${dash}" 2>&1 | jq '.dashboard' > dashboards/$(echo ${dash}|sed 's,db/,,g')-dashboard.json
  echo "backing up $dash"
done