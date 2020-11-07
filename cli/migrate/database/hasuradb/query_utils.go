package hasuradb

// DatabaseDriver for the different db drivers that can be used via the different data sources
type DatabaseDriver string

// V2Query is for all queries that can run on v2/query
type V2Query string

// V1Metadata is for all queries that can run on v1/metadata
type V1Metadata string

// DataSourcesEndpoints has all the new endpoints related to the data sources changes
type DataSourcesEndpoints string

// All DatabaseDriver possibilities
const (
	Postgres DatabaseDriver = "postgres"
	MySQL                   = "mysql"
	// TODO: add other database drivers here
)

// All allowed query types that are allowed on v2/query
const (
	RunSQLV2 V2Query = "run_sql"
	BulkV2           = "bulk"
	Select           = "select"
	Insert           = "insert"
	Update           = "update"
	Delete           = "delete"
	Count            = "count"
)

// All allowed query types that are allowed on v1/metadata
const (
	AddSource                   V1Metadata = "add_source"
	DropSource                             = "drop_source"
	ReloadSource                           = "reload_source"
	BulkMetadata                           = "bulk"
	TrackTable                             = "track_table"
	UnTrackTable                           = "untrack_table"
	SetTableIsEnum                         = "set_table_is_enum"
	SetTableCustomFields                   = "set_table_custom_fields"
	TrackFunction                          = "track_function"
	UnTrackFunction                        = "untrack_function"
	CreateObjectRelationship               = "create_object_relationship"
	CreateArrayRelationship                = "create_array_relationship"
	DropRelationship                       = "drop_relationship"
	SetRelationshipComment                 = "set_relationship_comment"
	RenameRelationship                     = "rename_relationship"
	AddComputedField                       = "add_computed_field"
	DropComputedField                      = "drop_computed_field"
	CreateRemoteRelationship               = "create_remote_relationship"
	UpdateRemoteRelationship               = "update_remote_relationship"
	DeleteRemoteRelationship               = "delete_remote_relationship"
	CreateInsertPermission                 = "create_insert_permission"
	CreateSelectPermission                 = "create_select_permission"
	CreateUpdatePermission                 = "create_update_permission"
	CreateDeletePermission                 = "create_delete_permission"
	DropInsertPermission                   = "drop_insert_permission"
	DropSelectPermission                   = "drop_select_permission"
	DropUpdatePermission                   = "drop_update_permission"
	DropDeletePermission                   = "drop_delete_permission"
	SetPermissionComment                   = "set_permission_comment"
	CreateEventTrigger                     = "create_event_trigger"
	DeleteEventTrigger                     = "delete_event_trigger"
	RedeliverEvent                         = "redeliver_event"
	InvokeEventTrigger                     = "invoke_event_trigger"
	GetInconsistentMetadata                = "get_inconsistent_metadata"
	DropInconsistentMetadata               = "drop_inconsistent_metadata"
	AddRemoteSchema                        = "add_remote_schema"
	RemoveRemoteSchema                     = "remove_remote_schema"
	ReloadRemoteSchema                     = "reload_remote_schema"
	IntrospectRemoteSchema                 = "introspect_remote_schema"
	CreateCronTrigger                      = "create_cron_trigger"
	DeleteCronTrigger                      = "delete_cron_trigger"
	CreateScheduledEvent                   = "create_scheduled_event"
	DeleteScheduledEvent                   = "delete_scheduled_event"
	GetScheduledEvents                     = "get_scheduled_events"
	GetEventInvocations                    = "get_event_invocations"
	CreateQueryCollection                  = "create_query_collection"
	DropQueryCollection                    = "drop_query_collection"
	AddQueryToCollection                   = "add_query_to_collection"
	DropQueryFromCollection                = "drop_query_from_collection"
	AddCollectionToAllowList               = "add_collection_to_allowlist"
	DropCollectionFromAllowList            = "drop_collection_from_allowlist"
	ReplaceMetadata                        = "replace_metadata"
	ExportMetadata                         = "export_metadata"
	ClearMetadata                          = "clear_metadata"
	ReloadMetadata                         = "reload_metadata"
	CreateAction                           = "create_action"
	DropAction                             = "drop_action"
	UpdateAction                           = "update_action"
	CreateActionPermission                 = "create_action_permission"
	DropActionPermission                   = "drop_action_permission"
	SetCustomTypes                         = "set_custom_types"
	DumpInternalState                      = "dump_internal_state"
	GetCatalogState                        = "get_catalog_state"
	SetCatalogState                        = "set_catalog_state"
)

// New endpoints for the new metadata changes
const (
	V2QueryEndpoint    DataSourcesEndpoints = "v2/query"
	V1MetadataEndpoint                      = "v1/metadata"
)

// GetV2Query helps construct a valid query that can be run on v2/query
// NOTE: `source` is manadatory for all queries to v2/query
func GetV2Query(queryType V2Query, args map[string]interface{}, source string) map[string]interface{} {
	return map[string]interface{}{
		"type":   queryType,
		"source": source,
		"args":   args,
	}
}

// GetV1MetadataQueryWithPrefix helps construct a valid query that can be run on v1/metadata
// NOTE: not all metadata queries require the source and the prefix.
func GetV1MetadataQueryWithPrefix(queryType V1Metadata, args map[string]interface{}, dbDriver DatabaseDriver, source string) map[string]interface{} {
	typePrefix := "pg_"
	if dbDriver != "" && dbDriver != Postgres {
		// Kept it simple since we don't have any other driver(s) atm.
		typePrefix = "mysq_"
	}

	requestType := typePrefix + string(queryType)

	if source == "" {
		return map[string]interface{}{
			"type": requestType,
			"args": args,
		}
	}

	return map[string]interface{}{
		"type":   requestType,
		"source": source,
		"args":   args,
	}
}

// GetV1MetadataQueryNoPrefix helps construct a valid query that can be run on v1/metadata
// NOTE: not all metadata queries require the source and the prefix.
func GetV1MetadataQueryNoPrefix(queryType V1Metadata, args map[string]interface{}, source string) map[string]interface{} {
	if source == "" {
		return map[string]interface{}{
			"type": queryType,
			"args": args,
		}
	}

	return map[string]interface{}{
		"type":   queryType,
		"source": source,
		"args":   args,
	}
}
