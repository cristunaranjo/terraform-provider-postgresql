resource "postgresql_default_privileges" "read_only_tables" {
  role     = "test_role"
  database = "test_db"
  schema   = "public"

  owner       = "db_owner"
  object_type = "table"
  privileges  = ["SELECT"]
}
