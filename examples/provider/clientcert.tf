provider "postgresql" {
  host     = "postgres_server_ip"
  port     = 5432
  database = "postgres"
  username = "postgres_user"
  password = "postgres_password"
  sslmode  = "require"
  clientcert {
    cert = "/path/to/public-certificate.pem"
    key  = "/path/to/private-key.pem"
  }
}
