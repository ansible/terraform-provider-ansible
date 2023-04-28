#!/bin/bash

set -eux

dir=$(pwd)
tempdir="$(mktemp -d $dir/temp.XXXXXX)"
export TF_CLI_CONFIG_FILE="$dir/ansible-dev.tfrc"

function teardown()
{
  rm -rf "$tempdir"
}

trap teardown EXIT

cat vault-decrypted.yml > vault-encrypted.yml
ansible-vault encrypt --vault-id testvault@vault_password vault-encrypted.yml

cp -v main.tf $tempdir
mv vault-encrypted.yml $tempdir
cp vault_password $tempdir

cd $tempdir

terraform init || true  # expected to fail
terraform apply --auto-approve
cat terraform.tfstate > ../actual_tfstate.json

cd ../../integration
set +e
go test -v
exit_code="$?"
set -e

exit "$exit_code"
