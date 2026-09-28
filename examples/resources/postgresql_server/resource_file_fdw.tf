resource "postgresql_extension" "ext_file_fdw" {
  name = "file_fdw"
}

resource "postgresql_server" "myserver_file" {
  server_name = "myserver_file"
  fdw_name    = "file_fdw"
  depends_on  = [postgresql_extension.ext_file_fdw]
}
