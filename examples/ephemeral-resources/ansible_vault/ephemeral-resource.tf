ephemeral "ansible_vault" "secret" {
  vault_file          = "${path.module}/secrets/db_password.yml"
  vault_password_file = "${path.module}/.vault_pass"
}

# Use the decrypted content in another resource without it appearing in state.
# Requires Terraform 1.10 or later.
resource "aws_secretsmanager_secret_version" "db" {
  secret_id     = aws_secretsmanager_secret.db.id
  secret_string = ephemeral.ansible_vault.secret.yaml
}
