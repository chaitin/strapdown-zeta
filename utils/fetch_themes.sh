#!/bin/bash
# Run the js in the index page to get the list
# $('.section-preview .container:first .preview h3').map(function(n,x){return $(x).text()})
themes=(Cerulean Cosmo Cyborg Darkly Flatly Journal Lumen Paper Readable Sandstone Simplex Slate Spacelab Superhero United Yeti)
for i in "${themes[@]}"; do
    i=`echo "$i" | tr 'A-Z' 'a-z'`
    echo "https://bootswatch.com/$i/bootstrap.min.css"
    curl -L  "https://bootswatch.com/$i/bootstrap.min.css" > "$i.min.css"
done

# curl "http://bootswatch.com/default/bootstrap-responsive.min.css" > bootstrap-responsive.min.css
# curl "http://bootswatch.com/default/bootstrap.min.css" > bootstrap.min.css
