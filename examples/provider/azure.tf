data "azurerm_client_config" "current" {
}

# https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/postgresql_flexible_server
resource "azurerm_postgresql_flexible_server" "pgsql" {
  # ...
  authentication {
    active_directory_auth_enabled = true
    password_auth_enabled         = false
    tenant_id                     = data.azurerm_client_config.current.tenant_id
  }
}

# https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/postgresql_flexible_server_active_directory_administrator
resource "azurerm_postgresql_flexible_server_active_directory_administrator" "administrators" {
  object_id           = "00000000-0000-0000-0000-000000000000"
  principal_name      = "Azure AD Admin Group"
  principal_type      = "Group"
  resource_group_name = var.rg_name
  server_name         = azurerm_postgresql_flexible_server.pgsql.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
}

provider "postgresql" {
  host                = azurerm_postgresql_flexible_server.pgsql.fqdn
  port                = 5432
  database            = "postgres"
  username            = azurerm_postgresql_flexible_server_active_directory_administrator.administrators.principal_name
  sslmode             = "require"
  azure_identity_auth = true
  azure_tenant_id     = data.azurerm_client_config.current.tenant_id
}
