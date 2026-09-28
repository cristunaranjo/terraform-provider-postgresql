provider "postgresql" {
  alias    = "pg1"
  host     = "postgres_server_ip1"
  username = "postgres_user1"
  password = "postgres_password1"
}

provider "postgresql" {
  alias    = "pg2"
  host     = "postgres_server_ip2"
  username = "postgres_user2"
  password = "postgres_password2"
}

resource "postgresql_database" "my_db1" {
  provider = postgresql.pg1
  name     = "my_db1"
}

resource "postgresql_database" "my_db2" {
  provider = postgresql.pg2
  name     = "my_db2"
}
