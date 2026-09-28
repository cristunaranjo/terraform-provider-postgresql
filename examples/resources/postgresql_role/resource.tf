resource "postgresql_role" "my_role" {
  name     = "my_role"
  login    = true
  password = "mypass"
}

resource "postgresql_role" "my_replication_role" {
  name             = "replication_role"
  replication      = true
  login            = true
  connection_limit = 5
  password         = "md5c98cbfeb6a347a47eb8e96cfb4c4b890"
}

# Example using write-only password (password not stored in state)
resource "postgresql_role" "secure_role" {
  name                = "secure_role"
  login               = true
  password_wo         = "secure_password_123"
  password_wo_version = "1"
}
