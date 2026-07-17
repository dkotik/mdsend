#!/bin/sh

cat << EOF
Content-type: application/json

[
  {
    "name": "first",
    "email": "first@testmail.json"
  },
  {
    "name": "second",
    "email": "second@testmail.json"
  },
  {
    "name": "third",
    "email": "third@testmail.json"
  },
  {
    "name": "fourth",
    "email": "fourth@testmail.json"
  }
]
EOF
