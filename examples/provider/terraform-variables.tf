variable "host" {
  default = "localhost"
}

variable "password" {
  default = "adm"
}

variable "port" {
  default = 55432
}

provider "postgresql" {
  host     = var.host
  port     = var.port
  password = var.password
  sslmode  = "disable"
}

resource "postgresql_database" "test" {
  name = "test"
}
