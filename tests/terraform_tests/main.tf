terraform {
  required_providers {
    ansible = {
      version = "~> 1.0.0"
      source  = "ansible/ansible"
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
    a_string    = local.decoded_vault_yaml.hello
    a_number    = local.decoded_vault_yaml.a_number
    a_list      = local.decoded_vault_yaml.a_list
    a_map       = {
      key_one = "value_one"
      key_two = "value_two"
    }
    a_nested    = {
      a_string = "nested_value"
      a_number = 99
      a_bool   = false
      a_list   = ["x", "y"]
      a_map    = {
        inner_key = "inner_value"
      }
    }
  }
}

resource "ansible_group" "group" {
  name      = "somegroup"
  children  = ["somechild"]
  variables = {
    hello  = "from group!"
    a_bool = true
    a_number = 42
    a_list = ["one", "two"]
    a_map    = {
      key_a = "value_a"
      key_b = "value_b"
    }
    a_nested = {
      a_string = "nested_value"
      a_number = 7
      a_bool   = true
      a_list   = ["a", "b"]
      a_map    = {
        inner_key = "inner_value"
      }
    }
  }
}
