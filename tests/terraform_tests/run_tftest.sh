#!/bin/bash

tempdir="$(mktemp -d ./temp.XXXXXX)"

set -eu

cat vault-decrypted.yml > vault-encrypted.yml
ansible-vault encrypt --vault-id testvault@vault_password vault-encrypted.yml

cp -v main.tf $tempdir
mv vault-encrypted.yml $tempdir
cp vault_password $tempdir

cd $tempdir

terraform init
terraform apply --auto-approve
cat terraform.tfstate > ../actual_tfstate.json

cd ../../integration
set +e
go test -v
exit_code="$?"
set -e

cd ../terraform_tests
rm -rf "$tempdir"

exit "$exit_code"
