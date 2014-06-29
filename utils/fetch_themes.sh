#!/bin/bash
while read line; do
    curl $line/bootstrap.min.css > `basename $line`.min.css
done <<< "http://bootswatch.com/2/amelia/
http://bootswatch.com/2/cerulean/
http://bootswatch.com/2/cosmo/
http://bootswatch.com/2/cyborg/
http://bootswatch.com/2/flatly/
http://bootswatch.com/2/journal/
http://bootswatch.com/2/readable/
http://bootswatch.com/2/simplex/
http://bootswatch.com/2/slate/
http://bootswatch.com/2/spacelab/
http://bootswatch.com/2/spruce/
http://bootswatch.com/2/superhero/
http://bootswatch.com/2/united/"

curl "http://bootswatch.com/2/default/bootstrap-responsive.min.css" > bootstrap-responsive.min.css
curl "http://bootswatch.com/2/default/bootstrap.min.css" > bootstrap.min.css
