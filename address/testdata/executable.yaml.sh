#!/bin/sh

cat << EOF
Content-type: application/yaml

- name: first
  email: first@testmail.yaml
- name: second
  email: second@testmail.yaml
- name: third
  email: third@testmail.yaml
- name: fourth
  email: fourth@testmail.yaml
# - "./recipients.json"
EOF
