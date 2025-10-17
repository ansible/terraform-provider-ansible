resource "ansible_inventory" "host" {
  path = "${path.module}/inventory.json"
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
