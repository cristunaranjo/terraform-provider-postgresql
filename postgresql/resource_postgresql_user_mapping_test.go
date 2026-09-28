package postgresql

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestGetUserMappingUserServerName(t *testing.T) {
	cases := []struct {
		name         string
		attributes   map[string]any
		id           string
		username     string
		serverName   string
		errorMessage string
	}{
		{
			name: "from state",
			attributes: map[string]any{
				userMappingUserNameAttr:   "john.doe",
				userMappingServerNameAttr: "srv",
			},
			id:         "ignored",
			username:   "john.doe",
			serverName: "srv",
		},
		{
			name:       "import",
			id:         "remote.myserver",
			username:   "remote",
			serverName: "myserver",
		},
		{
			name:         "import ambiguous dotted name",
			id:           "a.b.c",
			errorMessage: "user mapping ID a.b.c has not the expected format 'user_name.server_name'",
		},
		{
			name:         "import without dot",
			id:           "nodot",
			errorMessage: "user mapping ID nodot has not the expected format 'user_name.server_name'",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourcePostgreSQLUserMapping().Schema, c.attributes)
			d.SetId(c.id)

			username, serverName, err := getUserMappingUserServerName(d)
			if c.errorMessage != "" {
				if err == nil || !strings.Contains(err.Error(), c.errorMessage) {
					t.Fatalf("expected error containing %q, got %v", c.errorMessage, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if username != c.username || serverName != c.serverName {
				t.Fatalf("got (%q, %q), expected (%q, %q)", username, serverName, c.username, c.serverName)
			}
		})
	}
}

func TestAccPostgresqlUserMapping_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testCheckCompatibleVersion(t, featureServer)
			testSuperuserPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPostgresqlUserMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresqlUserMappingConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlUserMappingExists("postgresql_user_mapping.remote"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "server_name", "myserver_postgres"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "user_name", "remote"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "options.user", "admin"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "options.password", "pass"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.special_chars", "options.password", "pass=$*'"),
				),
			},
			{
				ResourceName:      "postgresql_user_mapping.remote",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "postgresql_user_mapping.special_chars",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPostgresqlUserMapping_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testCheckCompatibleVersion(t, featureServer)
			testSuperuserPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPostgresqlUserMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresqlUserMappingConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlServerExists("postgresql_user_mapping.remote"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "server_name", "myserver_postgres"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "options.password", "pass"),
				),
			},
			{
				Config: testAccPostgresqlUserMappingChanges2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlServerExists("postgresql_user_mapping.remote"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "options.password", "passUpdated"),
				),
			},
			{
				Config: testAccPostgresqlUserMappingChanges3,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlServerExists("postgresql_user_mapping.remote"),
					resource.TestCheckResourceAttr(
						"postgresql_user_mapping.remote", "options.%", "0"),
				),
			},
		},
	})
}

func checkUserMappingExists(txn *sql.Tx, username string, serverName string) (bool, error) {
	var _rez bool
	err := txn.QueryRow("SELECT TRUE FROM pg_user_mappings WHERE usename = $1 AND srvname = $2", username, serverName).Scan(&_rez)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, fmt.Errorf("error reading info about user mapping: %s", err)
	}

	return true, nil
}

func testAccCheckPostgresqlUserMappingDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "postgresql_user_mapping" {
			continue
		}

		txn, err := startTransaction(client, "")
		if err != nil {
			return err
		}
		defer deferredRollback(txn)

		splitted := strings.Split(rs.Primary.ID, ".")
		exists, err := checkUserMappingExists(txn, splitted[0], splitted[1])

		if err != nil {
			return fmt.Errorf("error checking user mapping %s", err)
		}

		if exists {
			return fmt.Errorf("User mapping still exists after destroy")
		}
	}

	return nil
}

func testAccCheckPostgresqlUserMappingExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		username, ok := rs.Primary.Attributes[userMappingUserNameAttr]
		if !ok {
			return fmt.Errorf("No Attribute for username is set")
		}

		serverName, ok := rs.Primary.Attributes[userMappingServerNameAttr]
		if !ok {
			return fmt.Errorf("No Attribute for server name is set")
		}

		client := testAccProvider.Meta().(*Client)
		txn, err := startTransaction(client, "")
		if err != nil {
			return err
		}
		defer deferredRollback(txn)

		exists, err := checkUserMappingExists(txn, username, serverName)

		if err != nil {
			return fmt.Errorf("error checking user mapping %s", err)
		}

		if !exists {
			return fmt.Errorf("User mapping not found")
		}

		return nil
	}
}

var testAccPostgresqlUserMappingConfig = `
resource "postgresql_extension" "ext_postgres_fdw" {
  name = "postgres_fdw"
}

resource "postgresql_server" "myserver_postgres" {
  server_name = "myserver_postgres"
  fdw_name    = "postgres_fdw"
  options = {
    host   = "foo"
    dbname = "foodb"
    port   = "5432"
  }

  depends_on = [postgresql_extension.ext_postgres_fdw]
}

resource "postgresql_role" "remote" {
  name = "remote"
}

resource "postgresql_user_mapping" "remote" {
  server_name = postgresql_server.myserver_postgres.server_name
  user_name   = postgresql_role.remote.name
  options = {
    user = "admin"
    password = "pass"
  }
}

resource "postgresql_role" "special" {
  name = "special"
}

resource "postgresql_user_mapping" "special_chars" {
  server_name = postgresql_server.myserver_postgres.server_name
  user_name   = postgresql_role.special.name
  options = {
	user = "admin"
	password = "pass=$*'"
  }
}
`

var testAccPostgresqlUserMappingChanges2 = `
resource "postgresql_extension" "ext_postgres_fdw" {
	name = "postgres_fdw"
  }
  
  resource "postgresql_server" "myserver_postgres" {
	server_name = "myserver_postgres"
	fdw_name    = "postgres_fdw"
	options = {
	  host   = "foo"
	  dbname = "foodb"
	  port   = "5432"
	}
  
	depends_on = [postgresql_extension.ext_postgres_fdw]
  }
  
  resource "postgresql_role" "remote" {
	name = "remote"
  }
  
  resource "postgresql_user_mapping" "remote" {
	server_name = postgresql_server.myserver_postgres.server_name
	user_name   = postgresql_role.remote.name
	options = {
	  user = "admin"
	  password = "passUpdated"
	}
  }
`

var testAccPostgresqlUserMappingChanges3 = `
resource "postgresql_extension" "ext_postgres_fdw" {
	name = "postgres_fdw"
  }
  
  resource "postgresql_server" "myserver_postgres" {
	server_name = "myserver_postgres"
	fdw_name    = "postgres_fdw"
	options = {
	  host   = "foo"
	  dbname = "foodb"
	  port   = "5432"
	}
  
	depends_on = [postgresql_extension.ext_postgres_fdw]
  }
  
  resource "postgresql_role" "remote" {
	name = "remote"
  }
  
  resource "postgresql_user_mapping" "remote" {
	server_name = postgresql_server.myserver_postgres.server_name
	user_name   = postgresql_role.remote.name
  }
`
