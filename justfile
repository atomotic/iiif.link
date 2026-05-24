default:
    just -l

get-tify:
    curl -o assets/tify.js https://cdn.jsdelivr.net/npm/tify@0.35.0/dist/tify.js
    curl -o assets/tify.css https://cdn.jsdelivr.net/npm/tify@0.35.0/dist/tify.css


migrate:
    sqlite3 iiiflink.db < init.sql

deploy:
    GOOS=linux GOARCH=amd64 go build -o iiif.link
    ssh nebula "sudo systemctl stop iiiflink"
    scp iiif.link nebula:/opt/iiif.link/iiif.link
    ssh nebula "sudo systemctl start iiiflink"
