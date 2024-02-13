# Terraform Provider for Ansible

The Terraform Provider for Ansible provides a more straightforward and robust means of executing Ansible automation from Terraform than local-exec. Paired with the inventory plugin in [the Ansible cloud.terraform collection](https://github.com/ansible-collections/cloud.terraform), users can run Ansible playbooks and roles on infrastructure provisioned by Terraform. The provider also includes integrated ansible-vault support.

This provider can be [found in the Terraform Registry here](https://registry.terraform.io/providers/ansible/ansible/latest).

For more details on using Terraform and Ansible together see these blog posts:

* [Terraforming clouds with Ansible](https://www.ansible.com/blog/terraforming-clouds-with-ansible)
* [Walking on Clouds with Ansible](https://www.ansible.com/blog/walking-on-clouds-with-ansible)
* [Providing Terraform with that Ansible Magic](https://www.ansible.com/blog/providing-terraform-with-that-ansible-magic)


## Requirements

- install Go: [official installation guide](https://go.dev/doc/install)
- install Terraform: [official installation guide](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli)
- install Ansible: [official installation guide](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)

## Installation for Local Development

Run `make`. This will build a `terraform-provider-ansible` binary in the top level of the project. To get Terraform to use this binary, configure the [development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) for the provider installation. The easiest way to do this will be to create a config file with the following contents:

```
provider_installation {
  dev_overrides {
    "ansible/ansible" = "/path/to/project/root"
  }

  direct {}
}
```

The `/path/to/project/root` should point to the location where you have cloned this repo, where the `terraform-provider-ansible` binary will be built. You can then set the `TF_CLI_CONFIG_FILE` environment variable to point to this config file, and Terraform will use the provider binary you just built.

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
make test

# GH actions locally
./act
```

### Examples
The [examples](./examples/) subdirectory contains a usage example for this provider.

## Release notes

See the [generated changelog](https://github.com/ansible/terraform-provider-ansible/tree/main/CHANGELOG.rst).

## Releasing

To release a new version of the provider:

1. Update the version number in https://github.com/ansible/terraform-provider-ansible/blob/main/examples/provider/provider.tf
2. Run `go generate` to regenerate docs
3. Run `antsibull-changelog release --version <version>` to release a new version of the project.
4. Commit changes
5. Push a new tag (this should trigger an automated release process to the Terraform Registry)
6. Verify the new version is published at https://registry.terraform.io/providers/ansible/ansible/latest

## Licensing

GNU General Public License v3.0. See [LICENSE](/LICENSE) for full text.
