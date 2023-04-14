## Run the examples

**NOTE:** to run this example, you must have installed:
- terraform
- python
- docker
- golang
- ansible
------------------------------

In base directory of this project ``terraform-provider-ansible/`` run (if not already):
```shell
make build-dev
```

In ``terraform-provider-ansible/examples/ansible_playbook`` directory
```shell
terraform init
terraform apply

# For terraform debug mode, use:
env TF_LOG=TRACE terraform apply

# To destroy terraform instance:
terraform destroy
```
------
**NOTE**: if ``terraform destroy`` for whatever reason fails, you can manually destroy it by 
deleting all terraform related files in this directory. You may also need to [manually stop/remove all
created dockers](#to-delete-all-created-dockers).

------
##  How to check if everything works correctly
Upon running ``terraform apply``, there should be three dockers generated:
- By a simple example``simple.tf``:
    - ``julia-the-first`` 
    - ``julia-the-second``
- By an end-to-end test ``end-to-end.tf``:
    - ``julia``

To check this, use:
```shell
docker ps
```

To connect to the julia dockers, use:
```shell
docker exec -it <julia_docker_name> /bin/sh
```

### Quick links:
1. [Check the succession of end-to-end tests](#expected-results-for-the-end-to-end-tests)
2. [Check the succession of the simple example](#expected-results-for-the-simple-example)

## Expected results for the end-to-end tests
On the ``julia`` docker, there should be 7 text files with a prefix ``test_e2e``, one for each ``e2e`` resource in
``end-to-end.tf`` (excluding ``e2e_limit_negative``, which should fail to run).

To check the contents of these files, use:
```shell
cat /test_e2e*
```

The output should look like this (the order of ``cat`` file outputs may not be the same for you):

*To check the differences faster, you may find [this](https://www.diffchecker.com/text-compare/) useful.*
```
----------
test_e2e_groups.txt
i have executed in tag1!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_groups.txt
i have executed in tag2!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_groups.txt
i have executed in a group!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_groups.txt
SHOULD EXECUTE IF NO TAG SPECIFIED: TAG NEVER SPECIFIED
----------
test_e2e_limit_positive.txt
i have executed in tag1!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_limit_positive.txt
i have executed in tag2!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_limit_positive.txt
i have executed in a group!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_limit_positive.txt
SHOULD EXECUTE IF NO TAG SPECIFIED: TAG NEVER SPECIFIED
----------
test_e2e_tags.txt
i have executed in tag1!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_tags.txt
i have executed in tag2!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_tags_1.txt
i have executed in tag1!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_tags_2.txt
i have executed in tag2!
var not injected
var file not specified
vault file not specified
----------
----------
test_e2e_vars.txt
i have executed in tag1!
content
content from a var file
vault file not specified
----------
----------
test_e2e_vars.txt
i have executed in tag2!
content
content from a var file
vault file not specified
----------
----------
test_e2e_vars.txt
i have executed in a group!
content
content from a var file
vault file not specified
----------
----------
test_e2e_vars.txt
SHOULD EXECUTE IF NO TAG SPECIFIED: TAG NEVER SPECIFIED
----------
test_e2e_vault.txt
i have executed in tag1!
var not injected
var file not specified
content from a vault file
----------
----------
test_e2e_vault.txt
i have executed in tag2!
var not injected
var file not specified
content from a vault file
----------
----------
test_e2e_vault.txt
i have executed in a group!
var not injected
var file not specified
content from a vault file
----------
----------
test_e2e_vault.txt
SHOULD EXECUTE IF NO TAG SPECIFIED: TAG NEVER SPECIFIED
```

## Expected results for the simple example
On both dockers, there should be a text file ``~/simple-file.txt``.
On ``julia-the-second``, this file should contain content of some variables from vaults and var files. Those variables are:
- ``content_from_a_vault_file`` → from ``./vault-file.yml``
- ``content_from_a_var_file``   → from ``./var-file.yml``

The content of ``~/simple-file.txt`` should be something like this:
```
Hello, World!
Hello
content from a var file
content from a vault file
```

On ``julia-the-first``, this file (``~/simple-file.txt``) should have no content.

## How to build a docker from Dockerfile manually (no Terraform)
In ``terraform-provider-ansible/examples/ansible_playbook``:
```shell
docker build -t docker_image .
docker run --rm -d --name docker_name docker_image sleep infinity
```

###  To delete all created dockers
```shell
docker stop $(docker ps -a -q) && docker container prune
```
