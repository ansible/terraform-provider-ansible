# Terraform Provider for Ansible

The Terraform Provider for Ansible provides a more straightforward and robust means of executing Ansible automation from Terraform than local-exec. Paired with the inventory plugin in [the Ansible cloud.terraform collection](https://github.com/ansible-collections/cloud.terraform), uses can run Ansible playbooks and roles on infrastructure provisioned by Terraform. The provider also includes integrated ansible-vault support. 

This provider can be [found in the Terraform Registry here](https://registry.terraform.io/providers/ansible/ansible/latest).

For more details on using Terraform and Ansible together see these blog posts:

* [Terraforming clouds with Ansible](https://www.ansible.com/blog/terraforming-clouds-with-ansible)
* [Walking on Clouds with Ansible](https://www.ansible.com/blog/walking-on-clouds-with-ansible)
* [Providing Terraform with that Ansible Magic](https://www.ansible.com/blog/providing-terraform-with-that-ansible-magic)


## Requirements 

- install Go: [official installation guide](https://go.dev/doc/install)
- install Terraform: [official installation guide](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli)
- install Ansible: [official installation guide](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)

## Installation to Terraform

1. Clone this repository to any location on your computer (or download source code)
2. Use the command below from ``<local-path-to-repository>/terraform-provider-ansible``

```shell
make build-dev
```

### Testing

Lint:

```shell
curl -L https://github.com/golangci/golangci-lint/releases/download/v1.50.1/golangci-lint-1.50.1-linux-amd64.tar.gz \
    | tar --wildcards -xzf - --strip-components 1 "**/golangci-lint"
curl -L https://github.com/nektos/act/releases/download/v0.2.34/act_Linux_x86_64.tar.gz \
    | tar -xzf - act

# linters
./golangci-lint run -v

# tests
cd tests/terraform_tests
./run_tftest.sh

# GH actions locally
./act
```

### Examples
The [examples](./examples/) subdirectory contains a usage example for this provider.
