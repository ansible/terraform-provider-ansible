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
