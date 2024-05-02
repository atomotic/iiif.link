default:
    just -l

get-tify:
    curl -o assets/tify.js https://cdn.jsdelivr.net/npm/tify@0.30.2/dist/tify.js
    curl -o assets/tify.css https://cdn.jsdelivr.net/npm/tify@0.30.2/dist/tify.css

migrate:
    sqlite3 iiiflink.db < init.sql