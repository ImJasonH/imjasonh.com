cd static/
gsutil cp index.html gs://www.imjasonh.com/
gsutil -h "Content-Type: text/html" cp resume gs://www.imjasonh.com/
gsutil -h "Content-Type: text/html" cp projects gs://www.imjasonh.com/
gsutil -h "Content-Type: text/html" cp cursor gs://www.imjasonh.com/
gsutil -h "Content-Type: text/html" cp dice gs://www.imjasonh.com/
cd -
