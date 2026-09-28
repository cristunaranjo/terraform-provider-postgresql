resource "postgresql_role" "bob" {
  name = "bob"

  lifecycle {
    ignore_changes = [
      roles,
    ]
  }
}

resource "postgresql_grant_role" "bob_admin" {
  role       = "bob"
  grant_role = "admin"
}
