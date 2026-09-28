resource "postgresql_role" "app_www" {
  name = "app_www"
}

resource "postgresql_role" "app_dba" {
  name = "app_dba"
}

resource "postgresql_role" "app_releng" {
  name = "app_releng"
}

resource "postgresql_schema" "my_schema" {
  name  = "my_schema"
  owner = "postgres"

  policy {
    usage = true
    role  = postgresql_role.app_www.name
  }

  # app_releng can create new objects in the schema.  This is the role that
  # migrations are executed as.
  policy {
    create = true
    usage  = true
    role   = postgresql_role.app_releng.name
  }

  policy {
    create_with_grant = true
    usage_with_grant  = true
    role              = postgresql_role.app_dba.name
  }
}
