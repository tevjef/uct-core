#!/bin/bash -eu

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 path-to-dashboard.json"
    exit 1
fi

dashboardjson=$1

cat <<EOF
{
  "dashboard":
EOF

cat $dashboardjson

cat <<EOF
,
  "inputs": [
    {
      "name": "DS_PROMETHEUS",
      "pluginId": "prometheus",
      "type": "datasource",
      "value": "prometheus"
    }
  ],
  "overwrite": true
}
EOF

