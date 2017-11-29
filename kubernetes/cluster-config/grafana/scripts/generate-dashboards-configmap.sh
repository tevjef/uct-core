#!/bin/bash

cat <<-EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: monitoring
data:
EOF

for f in dashboards/*-dashboard.json
do
  echo "  $(basename $f): |+"
  scripts/wrap-dashboard.sh $f | sed "s/^/    /g"
done

for f in datasources/*-datasource.json
do
  echo "  $(basename $f): |+"
  cat $f | sed "s/^/    /g"
done
