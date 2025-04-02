package access_log

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/turbot/tailpipe-plugin-sdk/formats"
	"github.com/turbot/tailpipe-plugin-sdk/mappers"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

var apacheRegexMap = map[string]string{
	`%%`:             `%`,                                                                                   // literal %
	`%a`:             `(?P<remote_addr>[^ ]*)`,                                                              // remote_addr as IP
	`%{c}a`:          `(?P<remote_addr>[^ ]*)`,                                                              // remote_addr as IP (underlying connection)
	`%A`:             `(?P<local_addr>[^ ]*)`,                                                               // local_addr as IP
	`%b`:             `(?P<body_bytes_sent>[^ ]*)`,                                                          // body_bytes_sent (- if no bytes sent)
	`%B`:             `(?P<body_bytes_sent>[^ ]*)`,                                                          // body_bytes_sent (0 if no bytes sent)
	`%D`:             `(?P<request_time_us>[^ ]*)`,                                                          // request_time in microseconds
	`%f`:             `(?P<filename>[^ ]*)`,                                                                 // filename
	`%h`:             `(?P<remote_addr>[^ ]*)`,                                                              // remote_addr as hostname (or IP if hostname unknown)
	`%{c}h`:          `(?P<remote_addr>[^ ]*)`,                                                              // remote_addr as hostname (or IP if hostname unknown) (underlying connection)
	`%H`:             `(?P<server_protocol>[^ ]*)`,                                                          // server_protocol
	`%k`:             `(?P<keepalive_requests>[^ ]*)`,                                                       // keepalive_requests
	`%l`:             `(?P<remote_logname>[^ ]*)`,                                                           // response from ident on client machine, almost always `-` (unknown)
	`%m`:             `(?P<request_method>[^ ]*)`,                                                           // request_method
	`%p`:             `(?P<server_port>[^ ]*)`,                                                              // server_port
	`%{canonical}p`:  `(?P<server_port>[^ ]*)`,                                                              // server_port (canonical)
	`%{local}p`:      `(?P<apache_port>[^ ]*)`,                                                              // apache_port (local) - port apache is bound on
	`%{remote}p`:     `(?P<client_port>[^ ]*)`,                                                              // client_port (remote)
	`%P`:             `(?P<pid>[^ ]*)`,                                                                      // pid
	`%{pid}P`:        `(?P<pid>[^ ]*)`,                                                                      // pid
	`%{tid}P`:        `(?P<thread_id>[^ ]*)`,                                                                // thread id
	`%{hextid}P`:     `(?P<hex_thread_id>[^ ]*)`,                                                            // hex thread id
	`%q`:             `(?P<query_string>[^ ]*)`,                                                             // query_string
	`%r`:             `(?P<request_method>\S+)(?: +(?P<request_uri>[^ ]+))?(?: +(?P<server_protocol>\S+))?`, // request split into request_method, request_uri, and server_protocol
	`%R`:             `(?P<handler>[^ ]*)`,                                                                  // handler (mod_core, mod_cgi, etc.)
	`%s`:             `(?P<status>[^ ]*)`,                                                                   // status
	`%<s`:            `(?P<status>[^ ]*)`,                                                                   // status
	`%>s`:            `(?P<status>[^ ]*)`,                                                                   // status (final)
	`%t`:             `\[(?P<timestamp>[^\]]*)\]`,                                                           // time_local
	`%T`:             `(?P<request_time>[^ ]*)`,                                                             // request_time in seconds
	`%{s}T`:          `(?P<request_time>[^ ]*)`,                                                             // request_time in seconds same as %T
	`%{ms}T`:         `(?P<request_time_ms>[^ ]*)`,                                                          // request_time in milliseconds
	`%{us}T`:         `(?P<request_time_us>[^ ]*)`,                                                          // request_time in microseconds (same as %D)
	`%u`:             `(?P<remote_user>[^ ]*)`,                                                              // remote_user
	`%<u`:            `(?P<remote_user>[^ ]*)`,                                                              // remote_user (same as %u)
	`%>u`:            `(?P<remote_user>[^ ]*)`,                                                              // remote_user (final)
	`%U`:             `(?P<request_uri>[^ ]*)`,                                                              // uri
	`%v`:             `(?P<server_name>[^ ]*)`,                                                              // server_name
	`%V`:             `(?P<server_name>[^ ]*)`,                                                              // server_name
	`%X`:             `(?P<connection_status>[^ ]*)`,                                                        // connection_status (x = connection aborted, + = connection may be kept alive, - = connection will be closed)
	`%I`:             `(?P<bytes_received>[^ ]*)`,                                                           // bytes_received
	`%O`:             `(?P<bytes_sent>[^ ]*)`,                                                               // bytes_sent
	`%S`:             `(?P<bytes_transferred>[^ ]*)`,                                                        // bytes sent and received
	`%{Referer}i`:    `(?P<http_referer>[^"]*)`,                                                             // Referer
	`%{User-agent}i`: `(?P<http_user_agent>[^"]*)`,                                                          // User-agent (linux, macOS)
	`%{User-Agent}i`: `(?P<http_user_agent>[^"]*)`,                                                          // User-Agent (Windows)
}

type AccessLogTableFormat struct {
	// the name of this format instance
	Name string `hcl:"name,label"`
	// Description of the format
	Description string `hcl:"description,optional"`
	// the layout of the log line
	Layout string `hcl:"layout"`
}

func NewAccessLogTableFormat() formats.Format {
	return &AccessLogTableFormat{}
}

func (a *AccessLogTableFormat) Validate() error {
	return nil
}

// Identifier returns the format TYPE
func (a *AccessLogTableFormat) Identifier() string {
	// format name is same as table name
	return AccessLogTableIdentifier
}

// GetName returns the format instance name
func (a *AccessLogTableFormat) GetName() string {
	// format name is same as table name
	return a.Name
}

// SetName sets the name of this format instance
func (a *AccessLogTableFormat) SetName(name string) {
	a.Name = name
}

func (a *AccessLogTableFormat) GetDescription() string {
	return a.Description
}

func (a *AccessLogTableFormat) GetMapper() (mappers.Mapper[*types.DynamicRow], error) {
	// convert the layout to a regex
	regex, err := a.GetRegex()
	if err != nil {
		return nil, err
	}
	return mappers.NewRegexMapper[*types.DynamicRow](regex)
}

// GetRegex converts the layout to a regex
func (a *AccessLogTableFormat) GetRegex() (string, error) {
	logFormat := a.Layout

	// extract time format contents
	timePattern := `%{([^}]+)}t`
	timeRegex := regexp.MustCompile(timePattern)
	timeLocations := timeRegex.FindAllStringSubmatchIndex(logFormat, -1)

	// extract Apache tokens
	tokenRegex := regexp.MustCompile(`%(?:[a-zA-Z]|[<>][a-zA-Z]|\{[^}]+\}[a-zA-Z])`)
	tokenLocations := tokenRegex.FindAllStringIndex(logFormat, -1)

	// Create a map of positions we want to preserve (not escape)
	preserveRanges := make(map[int]bool)

	// preserve time format ranges
	for _, loc := range timeLocations {
		if len(loc) >= 4 { // Each match has at least 4 indices (overall match start/end + group start/end)
			start, end := loc[0], loc[1]
			for j := start; j < end; j++ {
				preserveRanges[j] = true
			}
		}
	}

	// preserve token ranges
	for _, loc := range tokenLocations {
		for i := loc[0]; i < loc[1]; i++ {
			preserveRanges[i] = true
		}
	}

	// escape using regexp.QuoteMeta and preserve ranges that we don't want to escape (time formats and tokens)
	var result strings.Builder
	for i := 0; i < len(logFormat); i++ {
		if preserveRanges[i] {
			result.WriteByte(logFormat[i])
		} else {
			result.WriteString(regexp.QuoteMeta(string(logFormat[i])))
		}
	}
	logFormat = result.String()

	// replace time formats
	timeMatches := timeRegex.FindStringSubmatch(logFormat)
	if len(timeMatches) > 1 {
		timeFormat := timeFormatToRegex(timeMatches[1])
		logFormat = strings.Replace(logFormat, timeMatches[0], fmt.Sprintf(`(?P<timestamp>%s)`, timeFormat), 1)
	}

	// replace tokens with regex patterns
	tokens := tokenRegex.FindAllString(logFormat, -1)
	for _, token := range tokens {
		if regexValue, exists := apacheRegexMap[token]; exists {
			logFormat = strings.ReplaceAll(logFormat, token, regexValue)
		} else {
			return "", fmt.Errorf("unsupported token in format: %s", token)
		}
	}

	if logFormat != "" {
		logFormat = fmt.Sprintf("^%s", logFormat)
	}

	return logFormat, nil
}

func (a *AccessLogTableFormat) GetProperties() map[string]string {
	return map[string]string{
		"layout": a.Layout,
	}
}

// timeFormatToRegex converts a strftime time format to a regex
func timeFormatToRegex(format string) string {
	replacements := map[string]string{
		"%a": `[A-Za-z]+`,
		"%A": `[A-Za-z]+`,
		"%b": `[A-Za-z]{3}`,
		"%B": `[A-Za-z]+`,
		"%c": `.+`,
		"%d": `\d{2}`,
		"%e": `\d{1,2}`,
		"%f": `\d{6}`,
		"%H": `\d{2}`,
		"%I": `\d{2}`,
		"%j": `\d{3}`,
		"%m": `\d{2}`,
		"%M": `\d{2}`,
		"%p": `[APM]+`,
		"%S": `\d{2}`,
		"%U": `\d{2}`,
		"%w": `\d`,
		"%W": `\d{2}`,
		"%x": `.+`,
		"%X": `.+`,
		"%y": `\d{2}`,
		"%Y": `\d{4}`,
		"%z": `[+-]\d{4}`,
		"%Z": `[A-Za-z]+`,
	}

	for k, v := range replacements {
		format = strings.ReplaceAll(format, k, v)
	}

	return format
}
