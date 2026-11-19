Terraform Provider for PostgreSQL
=================================

This provider manages [PostgreSQL](https://www.postgresql.org/) objects such as databases, roles, grants, schemas and extensions with Terraform or OpenTofu.

It is a fork of [cyrilgdn/terraform-provider-postgresql](https://github.com/cyrilgdn/terraform-provider-postgresql), taken after upstream v1.27.0. Releases from v1.28.0 onwards are published as `cristunaranjo/postgresql`.

- Documentation: https://registry.terraform.io/providers/cristunaranjo/postgresql/latest/docs
- OpenTofu registry: https://search.opentofu.org/provider/cristunaranjo/postgresql

Changes from upstream
---------------------

- Destroying a `postgresql_grant` no longer fails when some of its tables, sequences or functions were dropped outside Terraform ([#1](https://github.com/cristunaranjo/terraform-provider-postgresql/pull/1), [#5](https://github.com/cristunaranjo/terraform-provider-postgresql/pull/5)).
- `terraform import` works for `postgresql_security_label`, `postgresql_server` and `postgresql_user_mapping` ([#11](https://github.com/cristunaranjo/terraform-provider-postgresql/pull/11)).
- `postgresql_server` options whose value contains `=` no longer show a permanent diff ([#18](https://github.com/cristunaranjo/terraform-provider-postgresql/pull/18)).
- Release binaries are built for amd64 and arm64 only ([#6](https://github.com/cristunaranjo/terraform-provider-postgresql/pull/6)). If you need 386 or 32-bit arm, stay on upstream.

See the [releases](https://github.com/cristunaranjo/terraform-provider-postgresql/releases) for the full list.

Moving from cyrilgdn/postgresql
-------------------------------

Change the provider source:

```hcl
terraform {
  required_providers {
    postgresql = {
      source = "cristunaranjo/postgresql"
    }
  }
}
```

Then point the existing state at the new provider and install it:

```sh
terraform state replace-provider cyrilgdn/postgresql cristunaranjo/postgresql
terraform init
terraform plan
```

The plan should show no changes. `replace-provider` only changes the current workspace, so run it once per workspace. Commit the updated `.terraform.lock.hcl`. With OpenTofu, run the same commands with `tofu`.

Requirements
------------

- [Terraform](https://developer.hashicorp.com/terraform/install) or [OpenTofu](https://opentofu.org/docs/intro/install/). CI runs the acceptance tests with Terraform 1.15.4 against PostgreSQL 12 to 18.
- [Go](https://go.dev/doc/install) 1.26, only to build the provider.

Building the provider
---------------------

```sh
git clone https://github.com/cristunaranjo/terraform-provider-postgresql
cd terraform-provider-postgresql
make build
```

`make build` runs `go install`, so the binary ends up in `$(go env GOPATH)/bin`.

Developing the provider
-----------------------

To generate or update documentation, run `make generate`. This needs Terraform on your `PATH` to format the examples; the schema export downloads Terraform 1.15.4 itself.

To run the unit tests:

```sh
make test
```

To run the full acceptance test suite against a local PostgreSQL in Docker:

```sh
make testacc
```

To run some acceptance tests by hand:

```sh
# start a local PostgreSQL container
make testacc_setup

# load the environment variables the tests need
source tests/switch_superuser.sh

# run the test(s) you are working on
TF_LOG=INFO go test -v ./postgresql -run ^TestAccPostgresqlRole_Basic$

# remove the container
make testacc_cleanup
```

Postgres and PostgreSQL are registered trademarks of the PostgreSQL Community Association of Canada. This project is not affiliated with or endorsed by the PostgreSQL project.
