resource "docker_container" "alpine_1" {
  image    = docker_image.julia.image_id
  name     = "julia"
  must_run = true

  command = [
    "sleep",
    "infinity"
  ]
}



# Test resources:
# - e2e-vars
# - e2e-vault
# - e2e-groups
# - e2e-limit-positive
# - e2e-limit-negative
# - e2e-tags


# NOTE: [ SUCCESS ]
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


# NOTE: [ SUCCESS ]
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

  depends_on = [ansible_playbook.e2e_vars] # make sure this resource waits for e2e_vars to finish
}


# NOTE: [ SUCCESS ]
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

  depends_on = [ansible_playbook.e2e_vault] # make sure this resource waits for e2e_vault to finish
}


# NOTE: [ SUCCESS ]
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

  depends_on = [ansible_playbook.e2e_groups] # make sure this resource waits for e2e_groups to finish
}


# NOTE: [ FAIL ]
#   -- this resource is supposed to fail,
#      so the playbook failure is being ignored
resource "ansible_playbook" "e2e_limit_negative" {
  ignore_playbook_failure = true  # set to 'true' because it's supposed to fail

  ansible_playbook_binary = "ansible-playbook"
  playbook                = "end-to-end-playbook.yml"

  # inventory configuration
  name = docker_container.alpine_1.name
  check_mode = true

  limit = [
    "nonexistent_host"
  ]

  # connection configuration and other vars
  extra_vars = {
    ansible_hostname   = docker_container.alpine_1.name
    ansible_connection = "docker"

    test_filename = "test_e2e_limit_negative.txt"
  }

  verbosity = 3

  depends_on = [ansible_playbook.e2e_limit_positive] # make sure this resource waits for e2e_limit_positive to finish
}


# NOTE: [ SUCCESS ]
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

  depends_on = [ansible_playbook.e2e_limit_negative] # make sure this resource waits for e2e_limit_negative to finish
}


# NOTE: [ SUCCESS ]
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

  depends_on = [ansible_playbook.e2e_tags] # make sure this resource waits for e2e_tags to finish
}


# NOTE: [ SUCCESS ]
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

  depends_on = [ansible_playbook.e2e_tags_1] # make sure this resource waits for e2e_tags_1 to finish
}
