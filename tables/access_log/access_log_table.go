package access_log

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
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

func (c *AccessLogTable) GetSupportedFormats() *formats.SupportedFormats {
	return &formats.SupportedFormats{
		Formats: map[string]func() formats.Format{
			AccessLogTableIdentifier:    NewAccessLogTableFormat,
			constants.SourceFormatRegex: formats.NewRegex,
		},
		DefaultFormat: DefaultApacheAccessLogFormat,
	}
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
				Type:       "VARCHAR",
			},
			{
				ColumnName: "remote_logname",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "remote_user",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "timestamp",
				Type:       "TIMESTAMP", // TODO: Confirm column type
			},
			{
				ColumnName: "request_method",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "request_uri",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "server_protocol",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "status",
				Type:       "INTEGER",
			},
			{
				ColumnName: "body_bytes_sent",
				Type:       "INTEGER",
			},
			{
				ColumnName: "http_referer",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "http_user_agent",
				Type:       "VARCHAR",
			},
			// additional fields
			{
				ColumnName: "local_addr",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "filename",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "keepalive_requests",
				Type:       "INTEGER",
			},
			{
				ColumnName: "server_name",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "server_port",
				Type:       "INTEGER",
			},
			{
				ColumnName: "client_port",
				Type:       "INTEGER",
			},
			{
				ColumnName: "apache_port",
				Type:       "INTEGER",
			},
			{
				ColumnName: "pid",
				Type:       "INTEGER",
			},
			{
				ColumnName: "thread_id",
				Type:       "INTEGER",
			},
			{
				ColumnName: "hex_thread_id",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "query_string",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "handler",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "request_time",
				Type:       "FLOAT",
			},
			{
				ColumnName: "request_time_ms",
				Type:       "FLOAT",
			},
			{
				ColumnName: "request_time_us",
				Type:       "FLOAT",
			},
			{
				ColumnName: "connection_status",
				Type:       "VARCHAR",
			},
			{
				ColumnName: "bytes_received",
				Type:       "INTEGER",
			},
			{
				ColumnName: "bytes_sent",
				Type:       "INTEGER",
			},
			{
				ColumnName: "bytes_transferred",
				Type:       "INTEGER",
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
			return nil, err
		}
		row.OutputColumns[constants.TpTimestamp] = t
	}

	// now call the base class to do the rest of the enrichment
	return c.CustomTableImpl.EnrichRow(row, sourceEnrichmentFields)
}
