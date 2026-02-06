resource "ansible_host" "host" {
  name   = "somehost"
  groups = ["somegroup"]

  variables = {
    greetings   = "from host!"
    some        = "variable"
    a_bool      = true
    a_number    = 3
    a_list      = ["web", "production"]
    a_map       = {
      key_one = "value_one"
      key_two = "value_two"
    }
  }
}
