resource "postgresql_publication" "publication" {
  name   = "publication"
  tables = ["public.test", "another_schema.test"]
}
