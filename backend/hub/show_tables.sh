#!/bin/bash

OUTFILE="model.sql"
DB="../../hub.db"

echo "-- Generated SQL schema from $DB" > "$OUTFILE"

for TABLE in $(sqlite3 "$DB" "SELECT name FROM sqlite_master WHERE type='table'"); do
  # echo "-- Table: $TABLE" >> "$OUTFILE" # not SQL, so commented
  sqlite3 "$DB" "SELECT sql FROM sqlite_master WHERE type='table' AND name='$TABLE'" | grep -v -e 'sql' -e '---' | awk '{printf "%s", $0}' | pg_format - --keyword-case 2 >> "$OUTFILE"
  echo >> "$OUTFILE"
done