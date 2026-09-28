# Initial password setup
resource "postgresql_role" "app_user" {
  name                = "app_user"
  login               = true
  password_wo         = "initial_password_123"
  password_wo_version = "1"
}

# To rotate the password, update both attributes:
# password_wo         = "new_password_456"
# password_wo_version = "2"
