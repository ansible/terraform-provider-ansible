resource "ansible_vault" "secrets" {
  vault_file          = "vault.yml"
  vault_password_file = "/path/to/file"
}
