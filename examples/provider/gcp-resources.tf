resource "google_sql_database_instance" "test" {
  project          = "test-project"
  name             = "test-instance"
  database_version = "POSTGRES_13"
  region           = "europe-west3"

  settings {
    tier = "db-f1-micro"
  }
}

resource "google_sql_user" "postgres" {
  project  = "test-project"
  name     = "postgres"
  instance = google_sql_database_instance.test.name
  password = "xxxxxxxx"
}


provider "postgresql" {
  scheme   = "gcppostgres"
  host     = google_sql_database_instance.test.connection_name
  username = google_sql_user.postgres.name
  password = google_sql_user.postgres.password
}

resource "postgresql_database" "test_db" {
  name = "test_db"
}
