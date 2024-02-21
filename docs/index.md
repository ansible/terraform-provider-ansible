---
page_title: "Ansible Provider"
subcategory: ""
description: |-
  Terraform provider for Ansible.
---

# Ansible Provider

The Ansible provider is used to interact with Ansible.

Use the navigation to the left to read about the available resources.


## Example Usage

```terraform
terraform {
  required_providers {
    ansible = {
      version = "~> 1.2.0"
      source  = "ansible/ansible"
    }
  }
}


resource "ansible_vault" "secrets" {
  vault_file          = "vault.yml"
  vault_password_file = "/path/to/file"
}


locals {
  decoded_vault_yaml = yamldecode(ansible_vault.secrets.yaml)
}

resource "ansible_host" "host" {
  name   = "somehost"
  groups = ["somegroup"]

  variables = {
    greetings   = "from host!"
    some        = "variable"
    yaml_hello  = local.decoded_vault_yaml.hello
    yaml_number = local.decoded_vault_yaml.a_number

    # using jsonencode() here is needed to stringify 
    # a list that looks like: [ element_1, element_2, ..., element_N ]
    yaml_list = jsonencode(local.decoded_vault_yaml.a_list)
  }
}

resource "ansible_group" "group" {
  name     = "somegroup"
  children = ["somechild"]
  variables = {
    hello = "from group!"
  }
}
```
