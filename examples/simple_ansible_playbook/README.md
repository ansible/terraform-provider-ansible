## Run this example

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
Upon running ``terraform apply``, there should be two dockers generated (``alpine-docker-1``, ``alpine-docker-2``).
To check this, use:
```shell
docker ps
```

On both dockers, there should be a text file ``~/example-play-file.txt``.
On ``alpine-docker-1``, this file should contain content of some variables from vaults and var files. Those variables are:
- ``dict``   → from ``./vault-1.yml``
- ``a_list`` → from ``./vault-2.yml``
- ``text``   → from ``var_file.yml``

The content of ``~/example-play-file.txt`` should be something like this:
```
-< {'a': 'ana', 'b': 'berta', 'c': 'cilka', 'd': 'dani'}, ['some', 'nice', 'list'], I am a var file
```

------
**NOTE**: To see the contents of vault files, they are already decrypted in files ``./vault-1-decrypted.yml`` and
``./vault-2-decrypted.yml``
------

On ``alpine-docker-2``, this file (``~/example-play-file.txt``) should have no content.

To connect to an alpine docker, use:
```shell
# x ... can be 1 or 2
docker exec -it alpine-docker-x /bin/sh
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
