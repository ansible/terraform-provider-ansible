action "ansible_playbook" "ansible" {
  config {
    playbooks            = ["${path.module}/playbook.yml"]
    inventory            = [ansible_inventory.host.path]
    ssh_private_key_file = "./ssh-private-key.pem"

    extra_vars = {
      var_a = "Some variable"
      var_b = "Another variable"
    }
  }
}
