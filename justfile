default:
    just -l

migrate:
    sqlite3 iiiflink.db < init.sql