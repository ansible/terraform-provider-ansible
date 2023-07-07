#!/bin/sh

set -eux

terraform init
terraform apply -auto-approve

ansible-playbook -i inventory.yml playbook.yml

ip=$(ansible-inventory -i inventory.yml --list | jq -r '.nginx.hosts[0]')
curl "http://${ip}" --fail
