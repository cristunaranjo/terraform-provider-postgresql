provider "postgresql" {
  scheme   = "gcppostgres"
  host     = "test-project/europe-west3/test-instance"
  username = "postgres"
  port     = 5432
  password = "test1234"

  superuser = false
}
