resource "postgresql_schema" "public" {
  name = "public"
}

resource "postgresql_schema" "schema_foo" {
  name  = "my_schema"
  owner = "postgres"

  policy {
    usage = true
  }
}
