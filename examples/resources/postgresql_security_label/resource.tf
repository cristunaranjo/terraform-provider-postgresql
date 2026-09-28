resource "postgresql_role" "my_role" {
  name  = "my_role"
  login = true
}

resource "postgresql_security_label" "workload" {
  object_type    = "role"
  object_name    = postgresql_role.my_role.name
  label_provider = "pgaadauth"
  label          = "aadauth,oid=00000000-0000-0000-0000-000000000000,type=service"
}
