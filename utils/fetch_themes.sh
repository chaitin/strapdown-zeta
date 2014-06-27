#!/bin/bash
while read line; do
    curl $line/bootstrap.min.css > `basename $line`.min.css
done <<< "http://bootswatch.com/amelia/
http://bootswatch.com/cerulean/
http://bootswatch.com/cosmo/
http://bootswatch.com/cyborg/
http://bootswatch.com/darkly/
http://bootswatch.com/flatly/
http://bootswatch.com/journal/
http://bootswatch.com/lumen/
http://bootswatch.com/readable/
http://bootswatch.com/simplex/
http://bootswatch.com/slate/
http://bootswatch.com/spacelab/
http://bootswatch.com/superhero/
http://bootswatch.com/united/
http://bootswatch.com/yeti/"

curl "http://bootswatch.com/bower_components/bootstrap/dist/css/bootstrap.css" > default.min.css
