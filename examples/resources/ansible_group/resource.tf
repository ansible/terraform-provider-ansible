resource "ansible_group" "group" {
  name     = "somegroup"
  children = ["somechild"]
  variables = {
    hello = "from group!"
  }
}
