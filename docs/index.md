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
      version = "~> 1.4.0"
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
    greetings = "from host!"
    some      = "variable"
    a_string  = local.decoded_vault_yaml.hello
    a_number  = local.decoded_vault_yaml.a_number
    a_list    = local.decoded_vault_yaml.a_list
    a_bool    = true
    a_map     = {
      key_one = "value_one"
      key_two = "value_two"
    }
  }
}

resource "ansible_group" "group" {
  name     = "somegroup"
  children = ["somechild"]
  variables = {
    hello    = "from group!"
    a_bool   = true
    a_number = 42
    a_list   = ["one", "two"]
    a_map    = {
      key_a = "value_a"
      key_b = "value_b"
    }
  }
}
```
