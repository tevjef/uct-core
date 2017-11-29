#!/bin/bash

mv production production-bak
cp -R staging production

find -E "production" -regex ".*\.(yml|yaml)" -print0 | xargs -0 sed -i '' 's/namespace: staging/namespace: production/g'
find -E "production" -regex ".*\.(yml|yaml)" -print0 | xargs -0 sed -i '' 's/:staging/:production/g'

# Clean up if there has been no errors
rm -r production-bak