provider "postgresql" {
  alias = "admindb"
}

resource "postgresql_database" "db1" {
  provider = postgresql.admindb

  name = "testdb1"
}
