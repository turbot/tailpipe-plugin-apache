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
				ColumnName:  "remote_addr",
				Description: "Client IP address that made the request",
				Type:        "varchar",
			},
			{
				ColumnName:  "remote_logname",
				Description: "Client logname from identd (if supplied)",
				Type:        "varchar",
			},
			{
				ColumnName:  "remote_user",
				Description: "Authenticated username of the client",
				Type:        "varchar",
			},
			{
				ColumnName:  "timestamp",
				Description: "Time when the request was received",
				Type:        "timestamp",
			},
			{
				ColumnName:  "request_method",
				Description: "HTTP method used in the request (GET, POST, etc.)",
				Type:        "varchar",
			},
			{
				ColumnName:  "request_uri",
				Description: "Full original request URI, including arguments",
				Type:        "varchar",
			},
			{
				ColumnName:  "server_protocol",
				Description: "Protocol and version used in the request (e.g., 'HTTP/1.1')",
				Type:        "varchar",
			},
			{
				ColumnName:  "status",
				Description: "HTTP response status code",
				Type:        "integer",
			},
			{
				ColumnName:  "body_bytes_sent",
				Description: "Number of bytes sent to the client, excluding headers",
				Type:        "integer",
			},
			{
				ColumnName:  "http_referer",
				Description: "Value of the 'Referer' request header",
				Type:        "varchar",
			},
			{
				ColumnName:  "http_user_agent",
				Description: "Value of the 'User-Agent' request header",
				Type:        "varchar",
			},
			// additional fields
			{
				ColumnName:  "local_addr",
				Description: "Local IP address that accepted the request",
				Type:        "varchar",
			},
			{
				ColumnName:  "filename",
				Description: "Filename of the requested resource",
				Type:        "varchar",
			},
			{
				ColumnName:  "keepalive_requests",
				Description: "Number of requests handled on this keepalive connection",
				Type:        "integer",
			},
			{
				ColumnName:  "server_name",
				Description: "Name of the server that processed the request",
				Type:        "varchar",
			},
			{
				ColumnName:  "server_port",
				Description: "Port number the server was listening on",
				Type:        "integer",
			},
			{
				ColumnName:  "client_port",
				Description: "Port number used by the client",
				Type:        "integer",
			},
			{
				ColumnName:  "apache_port",
				Description: "Port number Apache was listening on",
				Type:        "integer",
			},
			{
				ColumnName:  "pid",
				Description: "Process ID of the Apache worker that handled the request",
				Type:        "integer",
			},
			{
				ColumnName:  "thread_id",
				Description: "Thread ID that handled the request",
				Type:        "integer",
			},
			{
				ColumnName:  "hex_thread_id",
				Description: "Thread ID in hexadecimal format",
				Type:        "varchar",
			},
			{
				ColumnName:  "query_string",
				Description: "Query string portion of the request URI",
				Type:        "varchar",
			},
			{
				ColumnName:  "handler",
				Description: "Handler that generated the response",
				Type:        "varchar",
			},
			{
				ColumnName:  "request_time",
				Description: "Time taken to process the request in seconds",
				Type:        "float",
			},
			{
				ColumnName:  "request_time_ms",
				Description: "Time taken to process the request in milliseconds",
				Type:        "float",
			},
			{
				ColumnName:  "request_time_us",
				Description: "Time taken to process the request in microseconds",
				Type:        "float",
			},
			{
				ColumnName:  "connection_status",
				Description: "Final status of the connection",
				Type:        "varchar",
			},
			{
				ColumnName:  "bytes_received",
				Description: "Total number of bytes received from the client",
				Type:        "integer",
			},
			{
				ColumnName:  "bytes_sent",
				Description: "Total number of bytes sent to the client including headers",
				Type:        "integer",
			},
			{
				ColumnName:  "bytes_transferred",
				Description: "Total number of bytes transferred (sent + received)",
				Type:        "integer",
			},
		},
		NullIf: "-", // default null value
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
