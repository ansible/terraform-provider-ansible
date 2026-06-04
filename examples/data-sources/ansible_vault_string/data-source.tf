data "ansible_vault_string" "db_password" {
  content = <<-EOT
    $ANSIBLE_VAULT;1.1;AES256
    66386439653236336462626566653063336164663966303231363934653561363964613
    3562396563643434386566616637653564623436623437386237613438386231383164
    EOT
  vault_password_file = "${path.module}/.vault_pass"
}

output "db_password" {
  value     = data.ansible_vault_string.db_password.plaintext
  sensitive = true
}
