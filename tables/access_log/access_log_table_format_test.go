package access_log

import (
	"regexp"
	"testing"
)

func Test_AccessLogTableFormat_GetRegex(t *testing.T) {
	type args struct {
		layout  string
		logLine string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Default: common log format (CLF)",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b`,
				logLine: `192.168.1.1 - john [24/Feb/2025:12:34:56 +0000] "GET /index.html HTTP/1.1" 200 1234`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/index.html",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "1234",
			},
		},
		{
			name: "Default: combined log format",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-agent}i"`,
				logLine: `192.168.1.1 - john [24/Feb/2025:12:34:56 +0000] "GET /index.html HTTP/1.1" 200 1234 "https://example.com" "Turbot/Awesome"`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/index.html",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "1234",
				"http_referer":    "https://example.com",
				"http_user_agent": "Turbot/Awesome",
			},
		},
		{
			name: "Custom: Reordered fields",
			args: args{
				layout:  `%u %h %t %>s "%r" %b`,
				logLine: `john 192.168.1.1 [24/Feb/2025:12:34:56 +0000] 200 "GET /dashboard HTTP/1.1" 4321`,
			},
			want: map[string]string{
				"remote_user":     "john",
				"remote_addr":     "192.168.1.1",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"status":          "200",
				"request_method":  "GET",
				"request_uri":     "/dashboard",
				"server_protocol": "HTTP/1.1",
				"body_bytes_sent": "4321",
			},
		},
		{
			name: "Custom: ISO-8601 timestamp",
			args: args{
				layout:  `%h %l %u [%{%Y-%m-%dT%H:%M:%SZ}t] "%r" %>s %b`,
				logLine: `192.168.1.1 - john [2025-02-24T12:34:56Z] "GET /api/data HTTP/1.1" 201 9876`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "2025-02-24T12:34:56Z",
				"request_method":  "GET",
				"request_uri":     "/api/data",
				"server_protocol": "HTTP/1.1",
				"status":          "201",
				"body_bytes_sent": "9876",
			},
		},
		{
			name: "Custom: Referer and User-Agent omitted",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-Agent}i"`,
				logLine: `192.168.1.1 - john [24/Feb/2025:12:34:56 +0000] "GET /home HTTP/1.1" 304 - "" ""`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/home",
				"server_protocol": "HTTP/1.1",
				"status":          "304",
				"body_bytes_sent": "-",
				"http_referer":    "",
				"http_user_agent": "",
			},
		},
		{
			name: "Custom: Proxy logs with bytes received and sent",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b %I %O`,
				logLine: `192.168.1.1 - - [24/Feb/2025:12:34:56 +0000] "CONNECT example.com:443 HTTP/1.1" 200 0 1234 5678`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "-",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "CONNECT",
				"request_uri":     "example.com:443",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "0",
				"bytes_received":  "1234",
				"bytes_sent":      "5678",
			},
		},
		{
			name: "Unsupported token",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-agent}i" "%{X-RANDOM-IP}i"`,
				logLine: `192.168.1.1 - john [24/Feb/2025:12:34:56 +0000] "GET /data HTTP/1.1" 200 4321 "https://example.com" "Mozilla/5.0" "203.0.113.42"`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/data",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "4321",
				"http_referer":    "https://example.com",
				"http_user_agent": "Mozilla/5.0",
			},
			wantErr: true,
		},
		{
			name: "Custom: Different time zone format",
			args: args{
				layout:  `%h %l %u [%{%d/%b/%Y:%H:%M:%S %z}t] "%r" %>s %b`,
				logLine: `192.168.1.1 - - [24/Feb/2025:12:34:56 -0500] "POST /upload HTTP/1.1" 201 7890`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "-",
				"timestamp":       "24/Feb/2025:12:34:56 -0500",
				"request_method":  "POST",
				"request_uri":     "/upload",
				"server_protocol": "HTTP/1.1",
				"status":          "201",
				"body_bytes_sent": "7890",
			},
		},
		{
			name: "Custom: Tokens wrapped in double quotes",
			args: args{
				layout:  `"%h" "%l" "%u" "%t" "%r" "%>s" "%b"`,
				logLine: `"192.168.1.1" "-" "john" "[24/Feb/2025:12:34:56 +0000]" "GET /search?q=test HTTP/1.1" "200" "5123"`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/search?q=test",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "5123",
			},
		},
		{
			name: "Custom: Tokens wrapped in square brackets",
			args: args{
				layout:  `[%h] [%l] [%u] [%t] [%r] [%>s] [%b]`,
				logLine: `[192.168.1.1] [-] [john] [[24/Feb/2025:12:34:56 +0000]] [GET /admin HTTP/1.1] [403] [0]`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/admin",
				"server_protocol": "HTTP/1.1",
				"status":          "403",
				"body_bytes_sent": "0",
			},
		},
		{
			name: "Custom: Custom timestamp format without default brackets",
			args: args{
				layout:  `%h %l %u [%{%Y-%m-%d %H:%M:%S}t] "%r" %>s %b`,
				logLine: `192.168.1.1 - john [2025-02-24 12:34:56] "POST /api/v1/update HTTP/1.1" 201 6789`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "2025-02-24 12:34:56",
				"request_method":  "POST",
				"request_uri":     "/api/v1/update",
				"server_protocol": "HTTP/1.1",
				"status":          "201",
				"body_bytes_sent": "6789",
			},
		},
		{
			name: "Custom: Randomized Order and Mixed Wrappers",
			args: args{
				layout:  `[ %t ] ( %h ) { %u } "$%r$" << %>s >> $$%b$$`,
				logLine: `[ [24/Feb/2025:12:34:56 +0000] ] ( 192.168.1.1 ) { johndoe } "$PUT /config HTTP/1.1$" << 204 >> $$4321$$`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_user":     "johndoe",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "PUT",
				"request_uri":     "/config",
				"server_protocol": "HTTP/1.1",
				"status":          "204",
				"body_bytes_sent": "4321",
			},
		},
		{
			name: "Custom: Extended Time Format with Timezone Offset",
			args: args{
				layout:  `%h %l %u [%{%d/%b/%Y:%H:%M:%S %z}t] "%r" %>s %b`,
				logLine: `203.0.113.42 - - [24/Feb/2025:15:30:45 +0530] "PATCH /update-profile HTTP/1.1" 200 5678`,
			},
			want: map[string]string{
				"remote_addr":     "203.0.113.42",
				"remote_logname":  `-`,
				"remote_user":     "-",
				"timestamp":       "24/Feb/2025:15:30:45 +0530",
				"request_method":  "PATCH",
				"request_uri":     "/update-profile",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "5678",
			},
		},
		{
			name: "Custom: Missing Remote User Field",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b`,
				logLine: `192.168.1.1 - - [24/Feb/2025:12:34:56 +0000] "HEAD /ping HTTP/1.1" 200 -`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "-",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "HEAD",
				"request_uri":     "/ping",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "-",
			},
		},
		{
			name: "Custom: Request time in seconds (%T)",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b %T`,
				logLine: `192.168.1.1 - john [24/Feb/2025:12:34:56 +0000] "GET /data HTTP/1.1" 200 1234 3`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/data",
				"server_protocol": "HTTP/1.1",
				"status":          "200",
				"body_bytes_sent": "1234",
				"request_time":    "3", // %T: request time in seconds
			},
		},
		{
			name: "Custom: Request time in milliseconds (%{ms}T)",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b %{ms}T`,
				logLine: `192.168.1.1 - john [24/Feb/2025:12:34:56 +0000] "POST /api/update HTTP/1.1" 201 2048 135`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "john",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "POST",
				"request_uri":     "/api/update",
				"server_protocol": "HTTP/1.1",
				"status":          "201",
				"body_bytes_sent": "2048",
				"request_time_ms": "135", // %{ms}T: request time in milliseconds
			},
		},
		{
			name: "Custom: Server port (%p)",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b %p`,
				logLine: `192.168.1.1 - - [24/Feb/2025:12:34:56 +0000] "GET /home HTTP/2" 200 5123 443`,
			},
			want: map[string]string{
				"remote_addr":     "192.168.1.1",
				"remote_logname":  `-`,
				"remote_user":     "-",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "GET",
				"request_uri":     "/home",
				"server_protocol": "HTTP/2",
				"status":          "200",
				"body_bytes_sent": "5123",
				"server_port":     "443", // %p: server port
			},
		},
		{
			name: "Custom: Remote port (%{remote}p)",
			args: args{
				layout:  `%h %l %u %t "%r" %>s %b %{remote}p`,
				logLine: `203.0.113.42 - - [24/Feb/2025:12:34:56 +0000] "PUT /api/upload HTTP/1.1" 201 7890 54321`,
			},
			want: map[string]string{
				"remote_addr":     "203.0.113.42",
				"remote_logname":  `-`,
				"remote_user":     "-",
				"timestamp":       "24/Feb/2025:12:34:56 +0000",
				"request_method":  "PUT",
				"request_uri":     "/api/upload",
				"server_protocol": "HTTP/1.1",
				"status":          "201",
				"body_bytes_sent": "7890",
				"client_port":     "54321", // %{remote}p: client_port not server_port
			},
		},
	}

	for _, tt := range tests {
		format := &AccessLogTableFormat{
			Layout: tt.args.layout,
			Name:   "test",
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := format.getRegex()
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}

			// validate regex compiles
			re, err := regexp.Compile(got)
			if err != nil {
				t.Fatalf("error regex compile failed: %v", err)
			}

			// validate regex matches log line
			if !re.MatchString(tt.args.logLine) {
				t.Fatalf("error regex %s did not match log line: %v", re.String(), tt.args.logLine)
			}

			// validate named groups
			matches := re.FindStringSubmatch(tt.args.logLine)
			groups := make(map[string]string)
			for i, name := range re.SubexpNames() {
				if i != 0 && name != "" {
					groups[name] = matches[i]
				}
			}

			for k, v := range tt.want {
				if gotV, ok := groups[k]; ok {
					if gotV != v {
						t.Errorf("key %s: got %s, want %s", k, gotV, v)
					}
				} else {
					t.Errorf("key %s not found in matches", k)
				}
			}
		})
	}

}
