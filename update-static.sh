cd static/
gsutil -h "Content-Type: text/html" cp resume gs://www.imjasonh.com/
gsutil -h "Content-Type: text/html" cp projects gs://www.imjasonh.com/
gsutil cp -z ttf -R -a public-read . gs://www.imjasonh.com
cd -
