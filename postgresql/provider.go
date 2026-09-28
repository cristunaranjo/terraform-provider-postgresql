package postgresql

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/blang/semver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/oauth2/google"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
)

const (
	defaultProviderMaxOpenConnections            = 20
	defaultProviderConnMaxLifetimeSeconds        = 0 // unlimited
	defaultProviderMaxConnRetries                = 0
	defaultProviderConnectionRetryTimeoutSeconds = 5
	defaultExpectedPostgreSQLVersion             = "9.0.0"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"scheme": {
				Type:        schema.TypeString,
				Description: "The driver to use. Valid values are:\n  - `postgres`: Default value, use [`lib/pq`](https://pkg.go.dev/github.com/lib/pq)\n  - `awspostgres`: Use [GoCloud](#gocloud) for AWS\n  - `gcppostgres`: Use [GoCloud](#gocloud) for GCP",
				Optional:    true,
				Default:     "postgres",
				ValidateFunc: validation.StringInSlice([]string{
					"postgres",
					"awspostgres",
					"gcppostgres",
				}, false),
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PGHOST", nil),
				Description: "The address for the postgresql server connection, see [GoCloud](#gocloud) for specific format. Falls back to the `PGHOST` environment variable.",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PGPORT", 5432),
				Description: "The port for the postgresql server connection, or socket file name extension for Unix-domain connections. Falls back to the `PGPORT` environment variable, then to `5432`.",
			},
			"database": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Database to connect to. Falls back to the `PGDATABASE` environment variable, then to `postgres`.",
				DefaultFunc: schema.EnvDefaultFunc("PGDATABASE", "postgres"),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PGUSER", "postgres"),
				Description: "Username for the server connection. Falls back to the `PGUSER` environment variable, then to `postgres`.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PGPASSWORD", nil),
				Description: "Password for the server connection. Falls back to the `PGPASSWORD` environment variable.",
				Sensitive:   true,
			},

			"aws_rds_iam_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If set to `true`, call the AWS RDS API to grab a temporary password, using AWS Credentials from the environment (or the given profile, see `aws_rds_iam_profile`). See [IAM database authentication](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html).",
			},

			"aws_rds_iam_profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The AWS IAM Profile to use while using AWS RDS IAM Auth.",
			},

			"aws_rds_iam_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The AWS region to use while using AWS RDS IAM Auth.",
			},

			"aws_rds_iam_provider_role_arn": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "AWS IAM role to assume while using AWS RDS IAM Auth.",
			},

			"azure_identity_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If set to `true`, call the Azure OAuth token endpoint for temporary token, see [Azure](#azure).",
			},

			"azure_tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Azure tenant ID, required if `azure_identity_auth` is `true`. [Read more](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config.html).",
			},

			"gcp_iam_impersonate_service_account": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Service account to impersonate when using GCP IAM authentication, see [GCP](#gcp).",
			},

			// Connection username can be different than database username with user name maps (e.g.: in Azure)
			// See https://www.postgresql.org/docs/current/auth-username-maps.html
			"database_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username of the user in the database if different than connection username (See [user name maps](https://www.postgresql.org/docs/current/auth-username-maps.html)).",
			},

			"superuser": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PGSUPERUSER", true),
				Description: "Should be set to `false` if the user to connect is not a PostgreSQL superuser (as is the case in AWS RDS or GCP SQL). In this case, some features might be disabled (e.g.: Refreshing state password from database). Falls back to the `PGSUPERUSER` environment variable, then to `true`.",
			},

			"sslmode": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PGSSLMODE", nil),
				Description: "Set the priority for an SSL connection to the server. Falls back to the `PGSSLMODE` environment variable. Additional information on the options and their implications can be seen [in the `libpq(3)` SSL guide](http://www.postgresql.org/docs/current/static/libpq-ssl.html#LIBPQ-SSL-PROTECTION). Valid values for `sslmode` are (note: `prefer` is not supported by Go's [`lib/pq`](https://pkg.go.dev/github.com/lib/pq)):\n  - `disable` - No SSL\n  - `require` - Always SSL (the default, also skip verification)\n  - `verify-ca` - Always SSL (verify that the certificate presented by the server was signed by a trusted CA)\n  - `verify-full` - Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)",
			},
			"ssl_mode": {
				Type:        schema.TypeString,
				Description: "Deprecated alias of `sslmode`.",
				Optional:    true,
				Deprecated:  "Rename PostgreSQL provider `ssl_mode` attribute to `sslmode`",
			},
			"clientcert": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Configure the SSL client certificate.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cert": {
							Type:        schema.TypeString,
							Description: "The SSL client certificate file path. The file must contain PEM encoded data.",
							Required:    true,
						},
						"key": {
							Type:        schema.TypeString,
							Description: "The SSL client certificate private key file path. The file must contain PEM encoded data.",
							Required:    true,
						},
						"sslinline": {
							Type:        schema.TypeBool,
							Description: "If set to `true`, arguments accept inline ssl cert and key rather than a filename. Defaults to `false`.",
							Optional:    true,
						},
					},
				},
				MaxItems: 1,
			},
			"sslrootcert": {
				Type:        schema.TypeString,
				Description: "The SSL server root certificate file path. The file must contain PEM encoded data.",
				Optional:    true,
			},

			"connect_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("PGCONNECT_TIMEOUT", 180),
				Description:  "Maximum wait for connection, in seconds. Falls back to the `PGCONNECT_TIMEOUT` environment variable, then to `180`. Zero means wait indefinitely.",
				ValidateFunc: validation.IntAtLeast(-1),
			},
			"max_conn_retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      defaultProviderMaxConnRetries,
				Description:  "Maximum number of connection retries. Zero means no retries. The default is `0`.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"connection_retry_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      defaultProviderConnectionRetryTimeoutSeconds,
				Description:  "Maximum total wait, in seconds, across all connection retries. The default is `5`.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"max_connections": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      defaultProviderMaxOpenConnections,
				Description:  "Set the maximum number of open connections to the database. The default is `20`. Zero means unlimited open connections.",
				ValidateFunc: validation.IntAtLeast(-1),
			},
			"conn_max_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      defaultProviderConnMaxLifetimeSeconds,
				Description:  "Maximum lifetime of a connection, in seconds. The default is `0`. Zero means unlimited.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"expected_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultExpectedPostgreSQLVersion,
				Description:  "Specify a hint to Terraform regarding the expected version that the provider will be talking with. This is a required hint in order for Terraform to talk with an ancient version of PostgreSQL. This parameter is expected to be a [PostgreSQL Version](https://www.postgresql.org/support/versioning/) or `current`. Once a connection has been established, Terraform will fingerprint the actual version. Default: `9.0.0`.",
				ValidateFunc: validateExpectedVersion,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"postgresql_database":                  resourcePostgreSQLDatabase(),
			"postgresql_default_privileges":        resourcePostgreSQLDefaultPrivileges(),
			"postgresql_extension":                 resourcePostgreSQLExtension(),
			"postgresql_grant":                     resourcePostgreSQLGrant(),
			"postgresql_grant_role":                resourcePostgreSQLGrantRole(),
			"postgresql_replication_slot":          resourcePostgreSQLReplicationSlot(),
			"postgresql_publication":               resourcePostgreSQLPublication(),
			"postgresql_subscription":              resourcePostgreSQLSubscription(),
			"postgresql_physical_replication_slot": resourcePostgreSQLPhysicalReplicationSlot(),
			"postgresql_schema":                    resourcePostgreSQLSchema(),
			"postgresql_role":                      resourcePostgreSQLRole(),
			"postgresql_function":                  resourcePostgreSQLFunction(),
			"postgresql_server":                    resourcePostgreSQLServer(),
			"postgresql_user_mapping":              resourcePostgreSQLUserMapping(),
			"postgresql_security_label":            resourcePostgreSQLSecurityLabel(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"postgresql_schemas":   dataSourcePostgreSQLDatabaseSchemas(),
			"postgresql_tables":    dataSourcePostgreSQLDatabaseTables(),
			"postgresql_sequences": dataSourcePostgreSQLDatabaseSequences(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func validateExpectedVersion(v any, key string) (warnings []string, errors []error) {
	if _, err := semver.ParseTolerant(v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("invalid version (%q): %w", v.(string), err))
	}
	return
}

func getRDSAuthToken(region string, profile string, role string, username string, host string, port int) (string, error) {
	endpoint := fmt.Sprintf("%s:%d", host, port)

	ctx := context.Background()

	var awscfg aws.Config
	var err error

	if profile != "" {
		awscfg, err = awsConfig.LoadDefaultConfig(ctx, awsConfig.WithSharedConfigProfile(profile))
	} else if region != "" {
		awscfg, err = awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(region))
	} else {
		awscfg, err = awsConfig.LoadDefaultConfig(ctx)
	}
	if err != nil {
		return "", err
	}

	if role != "" {
		stsClient := sts.NewFromConfig(awscfg)
		roleInput := &sts.AssumeRoleInput{
			RoleArn:         aws.String(role),
			RoleSessionName: aws.String("TerraformPostgresqlProvider"),
		}

		roleOutput, err := stsClient.AssumeRole(ctx, roleInput)
		if err != nil {
			return "", fmt.Errorf("could not assume AWS role: %w", err)
		}

		awscfg, err = awsConfig.LoadDefaultConfig(ctx,
			awsConfig.WithCredentialsProvider(
				aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
					*roleOutput.Credentials.AccessKeyId,
					*roleOutput.Credentials.SecretAccessKey,
					*roleOutput.Credentials.SessionToken,
				)),
			),
		)
		if err != nil {
			return "", fmt.Errorf("could not load AWS default config: %w", err)
		}
	}

	token, err := auth.BuildAuthToken(ctx, endpoint, awscfg.Region, username, awscfg.Credentials)

	return token, err
}

func createGoogleCredsFileIfNeeded() error {
	if _, err := google.FindDefaultCredentials(context.Background()); err == nil {
		return nil
	}

	rawGoogleCredentials := os.Getenv("GOOGLE_CREDENTIALS")
	if rawGoogleCredentials == "" {
		return nil
	}

	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("could not create temporary file: %w", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			fmt.Printf("could not close temporary file: %v", err)
		}
	}()

	_, err = tmpFile.WriteString(rawGoogleCredentials)
	if err != nil {
		return fmt.Errorf("could not write in temporary file: %w", err)
	}

	return os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFile.Name())
}

func acquireAzureOauthToken(tenantId string) (string, error) {
	credential, err := azidentity.NewDefaultAzureCredential(
		&azidentity.DefaultAzureCredentialOptions{TenantID: tenantId})
	if err != nil {
		return "", err
	}
	token, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes:   []string{"https://ossrdbms-aad.database.windows.net/.default"},
		TenantID: tenantId,
	})
	if err != nil {
		return "", err
	}
	return token.Token, nil
}

func providerConfigure(d *schema.ResourceData) (any, error) {
	var sslMode string
	if sslModeRaw, ok := d.GetOk("sslmode"); ok {
		sslMode = sslModeRaw.(string)
	} else {
		sslModeDeprecated := d.Get("ssl_mode").(string)
		if sslModeDeprecated != "" {
			sslMode = sslModeDeprecated
		}
	}
	versionStr := d.Get("expected_version").(string)
	version, _ := semver.ParseTolerant(versionStr)

	host := d.Get("host").(string)
	port := d.Get("port").(int)
	username := d.Get("username").(string)

	var password string
	if d.Get("aws_rds_iam_auth").(bool) {
		profile := d.Get("aws_rds_iam_profile").(string)
		region := d.Get("aws_rds_iam_region").(string)
		role := d.Get("aws_rds_iam_provider_role_arn").(string)
		var err error
		password, err = getRDSAuthToken(region, profile, role, username, host, port)
		if err != nil {
			return nil, err
		}
	} else if d.Get("azure_identity_auth").(bool) {
		tenantId := d.Get("azure_tenant_id").(string)
		if tenantId == "" {
			return nil, fmt.Errorf("postgresql: azure_identity_auth is enabled, azure_tenant_id must be provided also")
		}
		var err error
		password, err = acquireAzureOauthToken(tenantId)
		if err != nil {
			return nil, err
		}
	} else {
		password = d.Get("password").(string)
	}

	config := Config{
		Scheme:                          d.Get("scheme").(string),
		Host:                            host,
		Port:                            port,
		Username:                        username,
		Password:                        password,
		DatabaseUsername:                d.Get("database_username").(string),
		Superuser:                       d.Get("superuser").(bool),
		SSLMode:                         sslMode,
		ApplicationName:                 "Terraform provider",
		ConnectTimeoutSec:               d.Get("connect_timeout").(int),
		MaxConnRetries:                  d.Get("max_conn_retries").(int),
		ConnectionRetryTimeoutSeconds:   d.Get("connection_retry_timeout_seconds").(int),
		MaxConns:                        d.Get("max_connections").(int),
		ConnMaxLifetimeSeconds:          d.Get("conn_max_lifetime_seconds").(int),
		ExpectedVersion:                 version,
		SSLRootCertPath:                 d.Get("sslrootcert").(string),
		GCPIAMImpersonateServiceAccount: d.Get("gcp_iam_impersonate_service_account").(string),
	}

	if value, ok := d.GetOk("clientcert"); ok {
		if spec, ok := value.([]any)[0].(map[string]interface{}); ok {
			config.SSLClientCert = &ClientCertificateConfig{
				CertificatePath: spec["cert"].(string),
				KeyPath:         spec["key"].(string),
				SSLInline:       spec["sslinline"].(bool),
			}
		}
	}

	if config.Scheme == "gcppostgres" {
		if err := createGoogleCredsFileIfNeeded(); err != nil {
			return nil, err
		}
	}

	client := config.NewClient(d.Get("database").(string))
	return client, nil
}
