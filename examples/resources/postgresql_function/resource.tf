resource "postgresql_function" "increment" {
  name = "increment"
  arg {
    name = "i"
    type = "integer"
  }
  returns  = "integer"
  language = "plpgsql"
  body     = <<-EOF
        BEGIN
            RETURN i + 1;
        END;
    EOF
}
