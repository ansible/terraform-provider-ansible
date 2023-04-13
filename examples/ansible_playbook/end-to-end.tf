terraform {
  required_providers {
    ansible = {
      source  = "ansible/ansible"
      version = "~> 1.0.0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.1"
    }
  }
}

resource "docker_image" "julia" {
  name = "julian-alps:latest"
  build {
    context    = "."
    dockerfile = "Dockerfile"
  }
}

resource "docker_container" "alpine_1" {
  image    = docker_image.julia.image_id
  name     = "julia"
  must_run = true

  command = [
    "sleep",
    "infinity"
  ]
}

#resource "docker_container" "alpine_2" {
#  image             = docker_image.julia.image_id
#  must_run          = true
#
#  command = [
#    "sleep",
#    "infinity"
#  ]
#}


# e2e-vars
# e2e-vault
# e2e-groups
# e2e-limit-positive
# e2e-limit-negative
# e2e-tags


resource "ansible_playbook" "e2e_vars" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name

  # play control
  var_files = [
    "var-file.yml"
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"
    injected_variable  = "content of an injected variable"

    test_filename = "test_e2e_vars.txt"
  }
}

# Checked: vault_id is optional, still not working
# TODO [done] : Check for other problems
#   --> "vault_password_file" parameter was accidentally renamed to "vault-password-file.txt"
#         --> had to also fix that in "provider/resource_vault.go" since the accidental change also affected that file
resource "ansible_playbook" "e2e_vault" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name

  # ansible vault
  vault_password_file = "vault-password-file.txt"
  vault_files = [
    "vault-file.yml",
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_vault.txt"
  }
}


resource "ansible_playbook" "e2e_groups" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name   = docker_container.alpine_1.name
  groups = ["this_group_exists"]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_groups.txt"
  }
}


resource "ansible_playbook" "e2e_limit_positive" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name

  limit = [
    docker_container.alpine_1.name
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_limit_positive.txt"
  }
}


resource "ansible_playbook" "e2e_limit_negative" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name

  limit = [
    "idonotexist"
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_limit_negative.txt"
  }
}


resource "ansible_playbook" "e2e_tags" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name

  tags = [
    "tag1",
    "tag2"
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_tags.txt"
  }
}


resource "ansible_playbook" "e2e_tags_1" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name

  tags = [
    "tag1"
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_tags_1.txt"
  }
}


resource "ansible_playbook" "e2e_tags_2" {
  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name

  tags = [
    "tag2"
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_tags_2.txt"
  }
}























##############################################
# SIMPLE.TF
# TODO: add simple.yml
##############################################
#
#resource "ansible_playbook" "example" {
#  ansible_playbook_binary = "ansible-playbook"
#  playbook                = "end-to-end-playbook.yml"
#
#  # inventory configuration
#  name   = docker_container.alpine_1.name
#  groups = ["playbook-group-1", "playbook-group-2"]
#
#  # ansible vault
#  vault_password_file = "vault_password_file"
#  vault_id            = "examplevault"
#  vault_files         = [
#    "vault-1.yml",
#    "vault-2.yml"
#  ]
#
#  # play control
#  tags = [
#    "tag1"
#  ]
#  limit = [
#    docker_container.alpine_1.name
#  ]
#  check_mode = false
#  diff_mode  = false
#  var_files  = [
#    "var_file.yml"
#  ]
#
#  # connection configuration and other vars
#  extra_vars = {
#    ansible_hostname   = docker_container.alpine_1.name
#    ansible_connection = "docker"
#  }
#
#  replayable = true
#  verbosity  = 3
#}
#
#resource "ansible_playbook" "example_2" {
#  playbook = "end-to-end-playbook.yml"
#  # inventory configuration
#  name     = docker_container.alpine_2.name
#  groups   = ["playbook-group-2"]
#
#  # ansible vault
#  vault_password_file = "vault_password_file"
#  vault_id            = "examplevault"
#  vault_files         = [
#    "vault-1.yml",
#    "vault-2.yml"
#  ]
#
#  # play control
#  tags = [
#    "tag2"
#  ]
#  limit = [
#    docker_container.alpine_2.name
#  ]
#  check_mode = false
#  diff_mode  = false
#  var_files  = [
#    "var_file.yml"
#  ]
#
#  # connection configuration and other vars
#  extra_vars = {
#    ansible_hostname   = docker_container.alpine_2.name
#    ansible_connection = "docker"
#    injected_var       = ""
#  }
#}
#
#resource "ansible_playbook" "example_2" {
#  playbook = "end-to-end-playbook.yml"
#  # inventory configuration
#  name     = docker_container.alpine_2.name
#  groups   = ["playbook-group-2"]
#
#  # ansible vault
#  vault_password_file = "vault_password_file"
#  vault_id            = "examplevault"
#  vault_files         = [
#    "vault-1.yml",
#    "vault-2.yml"
#  ]
#
#  # play control
#  tags = [
#    "tag2"
#  ]
#  limit = [
#    docker_container.alpine_2.name
#  ]
#  check_mode = false
#  diff_mode  = false
#  var_files  = [
#    "var_file.yml"
#  ]
#
#  # connection configuration and other vars
#  extra_vars = {
#    ansible_hostname   = docker_container.alpine_2.name
#    ansible_connection = "docker"
#    injected_var       = "SHOULD"
#  }
#}
