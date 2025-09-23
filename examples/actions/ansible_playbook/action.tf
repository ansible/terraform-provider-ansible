action "ansible_playbook" "ansible" {
  config {
    playbook             = "${path.module}/playbook.yml"
    name                 = "host-1.example.com"
    ssh_user             = "ubuntu"
    ssh_private_key_file = "./ssh-private-key.pem"

    extra_vars = {
      var_a = "Some variable"
      var_b = "Another variable"
    }
  }
}
