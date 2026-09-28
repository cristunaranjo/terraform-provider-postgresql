provider "postgresql" {
  scheme = "gcppostgres"
  host   = "test-project/europe-west3/test-instance"
  port   = 5432

  username                            = "service_account_id@$project_id.iam"
  gcp_iam_impersonate_service_account = "service_account_id@$project_id.iam.gserviceaccount.com"

  superuser = false
}
