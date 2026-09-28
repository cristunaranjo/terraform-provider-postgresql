package postgresql

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	tableQuery = `
	SELECT table_name, table_schema, table_type
	FROM information_schema.tables
	`
	tablePatternMatchingTarget = "table_name"
	tableSchemaKeyword         = "table_schema"
	tableTypeKeyword           = "table_type"
)

func dataSourcePostgreSQLDatabaseTables() *schema.Resource {
	return &schema.Resource{
		Read:        PGResourceFunc(dataSourcePostgreSQLTablesRead),
		Description: "The `postgresql_tables` data source retrieves a list of table names from a specified PostgreSQL database.",
		Schema: map[string]*schema.Schema{
			"database": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The PostgreSQL database which will be queried for table names.",
			},
			"schemas": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    0,
				Description: "List of PostgreSQL schema(s) which will be queried for table names. Queries all schemas in the database by default.",
			},
			"table_types": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    0,
				Description: "List of PostgreSQL table types which will be queried for table names. Includes all table types by default (including views and temp tables). Use 'BASE TABLE' for normal tables only.",
			},
			"like_any_patterns": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    0,
				Description: "List of expressions which will be pattern matched against table names in the query using the PostgreSQL `LIKE ANY` operators.",
			},
			"like_all_patterns": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    0,
				Description: "List of expressions which will be pattern matched against table names in the query using the PostgreSQL `LIKE ALL` operators.",
			},
			"not_like_all_patterns": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    0,
				Description: "List of expressions which will be pattern matched against table names in the query using the PostgreSQL `NOT LIKE ALL` operators.",
			},
			"regex_pattern": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expression which will be pattern matched against table names in the query using the PostgreSQL `~` (regular expression match) operator.",
			},
			"tables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"schema_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"table_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Description: "A list of PostgreSQL tables retrieved by this data source. Each table consists of `object_name` (the table name), `schema_name` (the parent schema) and `table_type` (the table type as defined in `information_schema.tables`).",
			},
		},
	}
}

func dataSourcePostgreSQLTablesRead(db *DBConnection, d *schema.ResourceData) error {
	database := d.Get("database").(string)

	txn, err := startTransaction(db.client, database)
	if err != nil {
		return err
	}
	defer deferredRollback(txn)

	query := tableQuery
	queryConcatKeyword := queryConcatKeywordWhere

	query = applyTableDataSourceQueryFilters(query, queryConcatKeyword, d)

	rows, err := txn.Query(query)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v\n", err)
		}
	}()

	tables := make([]any, 0)
	for rows.Next() {
		var object_name string
		var schema_name string
		var table_type string

		if err = rows.Scan(&object_name, &schema_name, &table_type); err != nil {
			return fmt.Errorf("could not scan table output for database: %w", err)
		}

		result := make(map[string]any)
		result["object_name"] = object_name
		result["schema_name"] = schema_name
		result["table_type"] = table_type
		tables = append(tables, result)
	}

	d.Set("tables", tables)
	d.SetId(generateDataSourceTablesID(d, database))

	return nil
}

func generateDataSourceTablesID(d *schema.ResourceData, databaseName string) string {
	return strings.Join([]string{
		databaseName,
		generatePatternArrayString(d.Get("schemas").([]any), queryArrayKeywordAny),
		generatePatternArrayString(d.Get("table_types").([]any), queryArrayKeywordAny),
		generatePatternArrayString(d.Get("like_any_patterns").([]any), queryArrayKeywordAny),
		generatePatternArrayString(d.Get("like_all_patterns").([]any), queryArrayKeywordAll),
		generatePatternArrayString(d.Get("not_like_all_patterns").([]any), queryArrayKeywordAll),
		d.Get("regex_pattern").(string),
	}, "_")
}

func applyTableDataSourceQueryFilters(query string, queryConcatKeyword string, d *schema.ResourceData) string {
	filters := []string{}
	schemasTypeFilter := applyTypeMatchingToQuery(tableSchemaKeyword, d.Get("schemas").([]any))
	if len(schemasTypeFilter) > 0 {
		filters = append(filters, schemasTypeFilter)
	}
	tableTypeFilter := applyTypeMatchingToQuery(tableTypeKeyword, d.Get("table_types").([]any))
	if len(tableTypeFilter) > 0 {
		filters = append(filters, tableTypeFilter)
	}
	filters = append(filters, applyPatternMatchingToQuery(tablePatternMatchingTarget, d)...)

	return finalizeQueryWithFilters(query, queryConcatKeyword, filters)
}
