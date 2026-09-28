resource "postgresql_replication_slot" "my_slot" {
  name   = "my_slot"
  plugin = "test_decoding"
}
