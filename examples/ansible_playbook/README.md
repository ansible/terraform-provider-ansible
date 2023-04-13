## Run the simple example

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
Upon running ``terraform apply``, there should be two dockers generated (``julia-the-first``, ``julia-the-second``).
To check this, use:
```shell
docker ps
```

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

To connect to the julia docker, use:
```shell
docker exec -it julia-the-[first | second] /bin/sh
```

## How to build a docker from Dockerfile manually (no Terraform)
In ``terraform-provider-ansible/examples/ansible_playbook``:
```shell
docker build -t docker_image .
docker run --rm -d --name docker_name docker_image sleep infinity
```

###  To delete all created dockers
```shell
docker stop $(docker ps -a -q)
docker container prune
```
