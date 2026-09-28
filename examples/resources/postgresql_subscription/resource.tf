resource "postgresql_subscription" "subscription" {
  name         = "subscription"
  conninfo     = "host=localhost port=5432 dbname=mydb user=postgres password=postgres"
  publications = ["publication"]
}
