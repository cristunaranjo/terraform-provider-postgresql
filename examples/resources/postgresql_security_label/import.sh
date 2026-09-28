# Security labels can be imported using `label_provider.object_type.object_name`. Quote the ID when the object type contains a space, e.g. "pgaadauth.materialized view.my_view".
terraform import postgresql_security_label.workload pgaadauth.role.my_role
