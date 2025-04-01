package access_log

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/error_types"
	"github.com/turbot/tailpipe-plugin-sdk/formats"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

const AccessLogTableIdentifier = "apache_access_log"
const AccessLogTableNilValue = "-"

// AccessLogTable - table for apache access logs
type AccessLogTable struct {
	table.CustomTableImpl
}

func (c *AccessLogTable) Identifier() string {
	return AccessLogTableIdentifier
}

func (c *AccessLogTable) GetDefaultFormat() formats.Format {
	return DefaultApacheAccessLogFormat
}

func (c *AccessLogTable) GetTableDefinition() *schema.TableSchema {
	return &schema.TableSchema{
		Name: AccessLogTableIdentifier,
		Columns: []*schema.ColumnSchema{
			{
				ColumnName: "tp_source_ip",
				SourceName: "remote_addr",
			},
			// default format fields
			{
				ColumnName: "remote_addr",
				Type:       "varchar",
			},
			{
				ColumnName: "remote_logname",
				Type:       "varchar",
			},
			{
				ColumnName: "remote_user",
				Type:       "varchar",
			},
			{
				ColumnName: "timestamp",
				Type:       "timestamp",
			},
			{
				ColumnName: "request_method",
				Type:       "varchar",
			},
			{
				ColumnName: "request_uri",
				Type:       "varchar",
			},
			{
				ColumnName: "server_protocol",
				Type:       "varchar",
			},
			{
				ColumnName: "status",
				Type:       "integer",
			},
			{
				ColumnName: "body_bytes_sent",
				Type:       "integer",
			},
			{
				ColumnName: "http_referer",
				Type:       "varchar",
			},
			{
				ColumnName: "http_user_agent",
				Type:       "varchar",
			},
			// additional fields
			{
				ColumnName: "local_addr",
				Type:       "varchar",
			},
			{
				ColumnName: "filename",
				Type:       "varchar",
			},
			{
				ColumnName: "keepalive_requests",
				Type:       "integer",
			},
			{
				ColumnName: "server_name",
				Type:       "varchar",
			},
			{
				ColumnName: "server_port",
				Type:       "integer",
			},
			{
				ColumnName: "client_port",
				Type:       "integer",
			},
			{
				ColumnName: "apache_port",
				Type:       "integer",
			},
			{
				ColumnName: "pid",
				Type:       "integer",
			},
			{
				ColumnName: "thread_id",
				Type:       "integer",
			},
			{
				ColumnName: "hex_thread_id",
				Type:       "varchar",
			},
			{
				ColumnName: "query_string",
				Type:       "varchar",
			},
			{
				ColumnName: "handler",
				Type:       "varchar",
			},
			{
				ColumnName: "request_time",
				Type:       "float",
			},
			{
				ColumnName: "request_time_ms",
				Type:       "float",
			},
			{
				ColumnName: "request_time_us",
				Type:       "float",
			},
			{
				ColumnName: "connection_status",
				Type:       "varchar",
			},
			{
				ColumnName: "bytes_received",
				Type:       "integer",
			},
			{
				ColumnName: "bytes_sent",
				Type:       "integer",
			},
			{
				ColumnName: "bytes_transferred",
				Type:       "integer",
			},
		},
		NullValue: "-", // default null value
	}
}

func (c *AccessLogTable) GetSourceMetadata() ([]*table.SourceMetadata[*types.DynamicRow], error) {
	// ask our CustomTableImpl for the mapper
	mapper, err := c.Format.GetMapper()
	if err != nil {
		return nil, err
	}

	// which source do we support?
	return []*table.SourceMetadata[*types.DynamicRow]{
		{
			// any artifact source
			SourceName: constants.ArtifactSourceIdentifier,
			Mapper:     mapper,
			Options: []row_source.RowSourceOption{
				artifact_source.WithRowPerLine(),
			},
		},
	}, nil
}

func (c *AccessLogTable) EnrichRow(row *types.DynamicRow, sourceEnrichmentFields schema.SourceEnrichment) (*types.DynamicRow, error) {
	if ts, ok := row.GetSourceValue("timestamp"); ok && ts != AccessLogTableNilValue {
		t, err := helpers.ParseTime(ts)
		if err != nil {
			return nil, error_types.NewRowErrorWithFields([]string{}, []string{"timestamp"})
		}
		row.OutputColumns[constants.TpTimestamp] = t
	}

	// tp_ips
	var ips []string
	if ip, ok := row.GetSourceValue("remote_addr"); ok && ip != AccessLogTableNilValue {
		ips = append(ips, ip)
	}
	if ip, ok := row.GetSourceValue("local_addr"); ok && ip != AccessLogTableNilValue {
		ips = append(ips, ip)
	}
	if len(ips) > 0 {
		row.OutputColumns[constants.TpIps] = ips
	}

	// tp_usernames
	if username, ok := row.GetSourceValue("remote_user"); ok && username != AccessLogTableNilValue {
		row.OutputColumns[constants.TpUsernames] = []string{username}
	}

	// now call the base class to do the rest of the enrichment
	return c.CustomTableImpl.EnrichRow(row, sourceEnrichmentFields)
}
