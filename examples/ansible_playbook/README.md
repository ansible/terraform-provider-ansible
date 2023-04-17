## Run the examples

**NOTE:** to run this example, you must have installed the following packages:

|       Name | Version used for testing  |
|-----------:|:--------------------------|
|  Terraform | v1.4.2                    |
|     Python | 3.10.6                    |
|     Docker | 23.0.1, build a5ee5b1     |
|     Golang | go1.18.10 linux/amd64     |
|    Ansible | core 2.14.3               |

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

```shell
# Save output of these files (sort the files alphabetically to make sure the output is always the same)
docker exec -it julia sh -c 'ls -X /test_e2e* | xargs cat' > end-to-end-actual-output

# Check diff of these files
diff end-to-end-expected-output end-to-end-actual-output
# if diff returns nothing, there are no differences.
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
