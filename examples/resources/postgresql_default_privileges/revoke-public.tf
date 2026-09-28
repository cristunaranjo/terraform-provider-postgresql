resource "postgresql_default_privileges" "revoke_public" {
  database    = postgresql_database.example_db.name
  role        = "public"
  owner       = "object_owner"
  object_type = "function"
  privileges  = []
}
