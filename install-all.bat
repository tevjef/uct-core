echo "Installing diff ..."
go install uct/common/uct-diff

echo "Installing clean ..."
go install uct/common/uct-clean

echo "Installing print..."
go install uct/common/uct-print

echo "Installing rutgers scraper ..."
go install uct/scrapers/rutgers


echo "Installing spiegal ..."
go install uct/servers/spiegal

echo "Installing db ..."
go install uct/db

echo "Installing gcm ..."
go install uct/gcm

echo "Installing influx-loggin ..."
go install uct/scripts/influx