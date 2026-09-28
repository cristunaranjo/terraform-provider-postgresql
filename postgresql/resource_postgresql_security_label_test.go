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

func TestGetSecurityLabelProviderTypeName(t *testing.T) {
	cases := []struct {
		name         string
		attributes   map[string]any
		id           string
		provider     string
		objectType   string
		objectName   string
		errorMessage string
	}{
		{
			name: "from state",
			attributes: map[string]any{
				securityLabelProviderAttr:   "dummy",
				securityLabelObjectTypeAttr: "role",
				securityLabelObjectNameAttr: "a.b",
			},
			id:         "ignored",
			provider:   "dummy",
			objectType: "role",
			objectName: "a.b",
		},
		{
			name:       "import",
			id:         "dummy.role.my_role",
			provider:   "dummy",
			objectType: "role",
			objectName: "my_role",
		},
		{
			name:       "import dotted object name",
			id:         "dummy.role.a.b.c",
			provider:   "dummy",
			objectType: "role",
			objectName: "a.b.c",
		},
		{
			name:       "import object type with space",
			id:         "dummy.materialized view.x",
			provider:   "dummy",
			objectType: "materialized view",
			objectName: "x",
		},
		{
			name:         "import malformed",
			id:           "dummy.role",
			errorMessage: "security label ID dummy.role has not the expected format 'label_provider.object_type.object_name'",
		},
		{
			name:       "import empty inner part of object name",
			id:         "dummy.role.a..b",
			provider:   "dummy",
			objectType: "role",
			objectName: "a..b",
		},
		{
			name:         "import empty object type",
			id:           "dummy..x",
			errorMessage: "security label ID dummy..x has not the expected format 'label_provider.object_type.object_name'",
		},
		{
			name:         "import empty provider",
			id:           ".role.x",
			errorMessage: "security label ID .role.x has not the expected format 'label_provider.object_type.object_name'",
		},
		{
			name:         "import empty object name",
			id:           "dummy.role.",
			errorMessage: "security label ID dummy.role. has not the expected format 'label_provider.object_type.object_name'",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourcePostgreSQLSecurityLabel().Schema, c.attributes)
			d.SetId(c.id)

			provider, objectType, objectName, err := getSecurityLabelProviderTypeName(d)
			if c.errorMessage != "" {
				if err == nil || !strings.Contains(err.Error(), c.errorMessage) {
					t.Fatalf("expected error containing %q, got %v", c.errorMessage, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if provider != c.provider || objectType != c.objectType || objectName != c.objectName {
				t.Fatalf("got (%q, %q, %q), expected (%q, %q, %q)", provider, objectType, objectName, c.provider, c.objectType, c.objectName)
			}
		})
	}
}

func TestAccPostgresqlSecurityLabel_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testCheckCompatibleVersion(t, featureSecurityLabel)
			testSuperuserPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPostgresqlSecurityLabelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresqlSecurityLabelConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlSecurityLabelExists("postgresql_security_label.test_label"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "object_type", "role"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "object_name", "security_label_test_role"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "label_provider", "dummy"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "label", "secret"),
				),
			},
			{
				ResourceName:      "postgresql_security_label.test_label",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPostgresqlSecurityLabel_ImportDottedObjectName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testCheckCompatibleVersion(t, featureSecurityLabel)
			testSuperuserPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPostgresqlSecurityLabelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresqlSecurityLabelDottedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlSecurityLabelExists("postgresql_security_label.test_label"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "id", "dummy.role.security_label.test.role"),
				),
			},
			{
				ResourceName:      "postgresql_security_label.test_label",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPostgresqlSecurityLabel_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testCheckCompatibleVersion(t, featureSecurityLabel)
			testSuperuserPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPostgresqlSecurityLabelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresqlSecurityLabelConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlSecurityLabelExists("postgresql_security_label.test_label"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "object_type", "role"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "object_name", "security_label_test_role"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "label_provider", "dummy"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "label", "secret"),
				),
			},
			{
				Config: testAccPostgresqlSecurityLabelChanges2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlSecurityLabelExists("postgresql_security_label.test_label"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "label", "top secret"),
				),
			},
			{
				Config: testAccPostgresqlSecurityLabelChanges3,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlSecurityLabelExists("postgresql_security_label.test_label"),
					resource.TestCheckResourceAttr(
						"postgresql_security_label.test_label", "object_name", "security_label_test-role2"),
				),
			},
		},
	})
}

func checkSecurityLabelExists(txn *sql.Tx, objectType string, objectName string, provider string) (bool, error) {
	var _rez bool
	err := txn.QueryRow("SELECT TRUE FROM pg_seclabels WHERE objtype = $1 AND objname = $2 AND provider = $3", objectType, quoteIdentifier(objectName), provider).Scan(&_rez)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, fmt.Errorf("error reading info about security label: %s", err)
	}

	return true, nil
}

func testAccCheckPostgresqlSecurityLabelDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "postgresql_security_label" {
			continue
		}

		txn, err := startTransaction(client, "")
		if err != nil {
			return err
		}
		defer deferredRollback(txn)

		splitted := strings.Split(rs.Primary.ID, ".")
		exists, err := checkSecurityLabelExists(txn, splitted[1], splitted[2], splitted[0])

		if err != nil {
			return fmt.Errorf("error checking security label%s", err)
		}

		if exists {
			return fmt.Errorf("Security label still exists after destroy")
		}
	}

	return nil
}

func testAccCheckPostgresqlSecurityLabelExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		objectType, ok := rs.Primary.Attributes[securityLabelObjectTypeAttr]
		if !ok {
			return fmt.Errorf("No Attribute for object type is set")
		}

		objectName, ok := rs.Primary.Attributes[securityLabelObjectNameAttr]
		if !ok {
			return fmt.Errorf("No Attribute for object name is set")
		}

		provider, ok := rs.Primary.Attributes[securityLabelProviderAttr]
		if !ok {
			return fmt.Errorf("No Attribute for security provider is set")
		}

		client := testAccProvider.Meta().(*Client)
		txn, err := startTransaction(client, "")
		if err != nil {
			return err
		}
		defer deferredRollback(txn)

		exists, err := checkSecurityLabelExists(txn, objectType, objectName, provider)

		if err != nil {
			return fmt.Errorf("error checking security label%s", err)
		}

		if !exists {
			return fmt.Errorf("Security label not found")
		}

		return nil
	}
}

var testAccPostgresqlSecurityLabelConfig = `
resource "postgresql_role" "test_role" {
  name            = "security_label_test_role"
  login           = true
  create_database = true
}
resource "postgresql_security_label" "test_label" {
  object_type    = "role"
  object_name    = postgresql_role.test_role.name
  label_provider = "dummy"
  label          = "secret"
}
`

var testAccPostgresqlSecurityLabelChanges2 = `
resource "postgresql_role" "test_role" {
  name            = "security_label_test_role"
  login           = true
  create_database = true
}
resource "postgresql_security_label" "test_label" {
  object_type    = "role"
  object_name    = postgresql_role.test_role.name
  label_provider = "dummy"
  label          = "top secret"
}
`

var testAccPostgresqlSecurityLabelChanges3 = `
resource "postgresql_role" "test_role" {
  name            = "security_label_test-role2"
  login           = true
  create_database = true
}
resource "postgresql_security_label" "test_label" {
  object_type    = "role"
  object_name    = postgresql_role.test_role.name
  label_provider = "dummy"
  label          = "top secret"
}
`

var testAccPostgresqlSecurityLabelDottedConfig = `
resource "postgresql_role" "test_role" {
  name = "security_label.test.role"
}
resource "postgresql_security_label" "test_label" {
  object_type    = "role"
  object_name    = postgresql_role.test_role.name
  label_provider = "dummy"
  label          = "secret"
}
`
