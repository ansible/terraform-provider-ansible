data "ansible_vault" "secret" {
  vault_file          = "${path.module}/secrets/db_password.yml"
  vault_password_file = "${path.module}/.vault_pass"
}

output "db_password_yaml" {
  value     = data.ansible_vault.secret.yaml
  sensitive = true
}
