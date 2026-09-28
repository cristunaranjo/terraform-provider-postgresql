provider "postgresql" {
  alias = "admindb"
}

resource "postgresql_role" "replication_role" {
  provider = postgresql.admindb

  name = "replication_name"
}
