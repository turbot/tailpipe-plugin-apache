package access_log

import (
	"github.com/turbot/tailpipe-plugin-sdk/formats"
)

var DefaultApacheAccessLogFormat = &formats.Regex{
	Name:        "apache_default",
	Description: "A default regex format that covers both Apache Common and Combined log formats.",
	Layout:      `^(?P<remote_addr>[^ ]*) (?P<remote_logname>[^ ]*) (?P<remote_user>[^ ]*) \[(?P<timestamp>[^\]]*)\] "(?P<request_method>\S+)(?: +(?P<request_uri>[^ ]+))?(?: +(?P<server_protocol>\S+))?" (?P<status>[^ ]*) (?P<body_bytes_sent>[^ ]*)(?: "(?P<http_referer>[^"]*)" "(?P<http_user_agent>[^"]*)")?$`,
}

var AccessLogTableFormatPresets = []formats.Format{
	&AccessLogTableFormat{
		Name:        "common",
		Description: "Apache Common Log Format",
		Layout:      `%h %l %u %t "%r" %>s %b`,
	},
	&AccessLogTableFormat{
		Name:        "combined",
		Description: "Apache Combined Log Format",
		Layout:      `%h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-agent}i"`,
	},
}
