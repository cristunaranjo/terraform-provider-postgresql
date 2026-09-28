resource "postgresql_default_privileges" "grant_table_privileges" {
  database    = postgresql_database.example_db.name
  role        = "current_role"
  owner       = "owner_role"
  schema      = "public"
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE"]
}
