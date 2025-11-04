data "ansible_inventory" "myinventory" {
  group {
    name = "webservers"

    host {
      name                     = aws_instance.web.public_ip
      ansible_user             = "ubuntu"
      ansible_private_key_file = local_file.private_key.filename
      ansible_ssh_extra_args   = "-o StrictHostKeyChecking=no"
    }
  }

  group {
    name = "dbservers"

    host {
      name                     = aws_instance.primary_db.public_ip
      ansible_user             = "root"
      ansible_private_key_file = local_file.private_key.filename
    }

    host {
      name                     = aws_instance.fallback_db.public_ip
      ansible_user             = "root"
      ansible_private_key_file = local_file.private_key.filename
    }
  }
}

# If you need the inventory as a file you can use the local_file resource
resource "local_file" "myinventory" {
  content  = ansible_inventory.myinventory.json
  filename = "${path.module}/inventory.json"
}

# It can also be used directly in an Action
action "ansible_playbook_run" "ansible" {
  config {
    playbooks   = ["${path.module}/playbook.yml"]
    inventories = [data.ansible_inventory.host.json]
  }
}
