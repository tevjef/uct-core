echo "Installing diff ..."
go install uct/common/uct-diff

echo "Installing clean ..."
go install uct/common/uct-clean

echo "Installing print..."
go install uct/common/uct-print

echo "Installing rutgers scraper ..."
go install uct/scrapers/rutgers

echo "Installing spike ..."
go install uct/servers/spike

echo "Installing db ..."
go install uct/db

echo "Installing hermes ..."
go install uct/hermes 

echo "Installing influx-loggin ..."
go install uct/scripts/influx