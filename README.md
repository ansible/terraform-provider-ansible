# Ansible Terraform Provider

_This template repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework). The template repository built on the [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk) can be found at [terraform-provider-scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding). See [Which SDK Should I Use?](https://developer.hashicorp.com/terraform/plugin/framework-benefits) in the Terraform documentation for additional information._

This repository is a *template* for a [Terraform](https://www.terraform.io) provider. It is intended as a starting point for creating Terraform providers, containing:

- A resource and a data source (`internal/provider/`),
- Examples (`examples/`) and generated documentation (`docs/`),
- Miscellaneous meta files.

These files contain boilerplate code that you will need to edit to create your own Terraform provider. Tutorials for creating Terraform providers can be found on the [HashiCorp Developer](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework) platform. _Terraform Plugin Framework specific guides are titled accordingly._


## Planned features
- Environment variables
- Ansible Configuration
- Better validation for hosts and variables
- Change the API to feel more like HashiCorp's providers
- Improve SSH Options handling

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19
- [Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html) >= 2.9

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
make build
```

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the root of the repository.

To generate or update documentation, run `make docs`.

In order to run the full suite of Acceptance tests, run `make testacc`.

You can run the 3 commands together:

```shell
make
```
