data "aws_secretsmanager_secret" "postgres_password" {
  name = "postgres_password"
}
data "aws_secretsmanager_secret_version" "postgres_password" {
  secret_id = data.aws_secretsmanager_secret.postgres_password.id
}

provider "postgresql" {
  # [...]
  password = jsondecode(data.aws_secretsmanager_secret_version.postgres_password.secret_string)["password"]
}
