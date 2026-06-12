data "ansible_inventory" "myinventory" {
  group {
    name = "webservers"

    host {
      name = "web1.example.com"
    }
  }
}

action "ansible_playbook_run" "with_inventories" {
  config {
    playbooks   = ["${path.module}/playbook.yml"]
    inventories = [data.ansible_inventory.myinventory.json]

    extra_vars = {
      var_a = "Some variable"
      var_b = "Another variable"
    }
  }
}

action "ansible_playbook_run" "with_inventory_files" {
  config {
    playbooks       = ["${path.module}/playbook.yml"]
    inventory_files = ["./hosts.ini", "./staging.yml"]

    private_key_file = "./ssh-private-key.pem"
  }
}
