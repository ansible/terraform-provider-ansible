terraform {
  required_providers {
    ansible = {
      version = "~> 0.0.2"
      source  = "terraform-ansible.com/ansibleprovider/ansible"
    }
  }
}


resource "ansible_vault" "secrets" {
  # required options
  vault_file          = "vault-encrypted.yml"
  vault_password_file = "vault_password"

  # optional options
  vault_id            = "testvault"
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
    yaml_list   = jsonencode(local.decoded_vault_yaml.a_list)
  }
}

resource "ansible_group" "group" {
  name      = "somegroup"
  children  = ["somechild"]
  variables = {
    hello = "from group!"
  }
}
