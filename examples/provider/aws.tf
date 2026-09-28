provider "postgresql" {
  scheme   = "awspostgres"
  host     = "test-instance.cvvrsv6scpgd.eu-central-1.rds.amazonaws.com"
  username = "postgres"
  port     = 5432
  password = "test1234"

  superuser = false
}
