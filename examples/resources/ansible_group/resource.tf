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
