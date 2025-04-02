---
title: "Tailpipe Table: apache_access_log - Query Apache Access Logs"
description: "Apache access logs capture detailed information about requests processed by the Apache HTTP server. This table provides a structured representation of the log data, including request details, client information, response codes, and processing times."
---

# Table: apache_access_log - Query Apache Access Logs

The `apache_access_log` table allows you to query Apache HTTP server access logs. This table provides detailed information about HTTP requests processed by your Apache servers, including client details, request information, response codes, and timing data.

By default, this table works with Apache's [combined log format](https://httpd.apache.org/docs/2.4/logs.html#combined):

```
LogFormat "%h %l %u %t \"%r\" %>s %b \"%{Referer}i\" \"%{User-agent}i\"" combined
```

Which contains the following fields:
- `%h` - Remote host (client IP)
- `%l` - Remote logname (from identd, if supplied)
- `%u` - Remote user (from auth)
- `%t` - Time the request was received
- `%r` - First line of request (method, URI, protocol)
- `%>s` - Status code
- `%b` - Size of response in bytes
- `%{Referer}i` - Referer header
- `%{User-agent}i` - User-Agent header

If your logs use a different format, you can specify a custom format as shown in the [example](https://hub.tailpipe.io/plugins/turbot/apache/tables/apache_access_log#collect-logs-with-custom-log-format) configurations below.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `apache_access_log`:

```sh
vi ~/.tailpipe/config/apache.tpc
```

```hcl
partition "apache_access_log" "my_apache_logs" {
  source "file" {
    paths       = ["/var/log/apache/access"]
    file_layout = `%{YEAR:year}%{MONTHNUM:month}%{MONTHDAY:day}_access.log`
  }
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) logs for all `apache_access_log` partitions:

```sh
tailpipe collect apache_access_log
```

Or for a single partition:

```sh
tailpipe collect apache_access_log.my_apache_logs
```

## Query

**[Explore example queries for this table →](https://hub.tailpipe.io/plugins/turbot/apache/queries/apache_access_log)**

### Failed Requests

Find failed HTTP requests (with status codes 400 and above) to troubleshoot server issues.

```sql
select
  timestamp,
  remote_addr,
  status,
  request_method,
  request_uri,
  server_protocol,
  body_bytes_sent
from
  apache_access_log
where
  status >= 400
order by
  tp_timestamp desc;
```

### Top 10 High Resource Usage Requests

Identify requests with high resource usage by analyzing bytes transferred and request times.

```sql
select
  timestamp,
  remote_addr,
  request_method,
  request_uri,
  bytes_transferred,
  request_time,
  status
from
  apache_access_log
where
  bytes_transferred > 1000000
  or request_time > 1.0
order by
  bytes_transferred desc,
  request_time desc
limit 10;
```

### Top 10 High Traffic Sources

Identify the IP addresses generating the most traffic.

```sql
select
  remote_addr,
  count(*) as request_count,
  count(distinct request_uri) as unique_urls,
  sum(body_bytes_sent) as total_bytes_sent
from
    apache_access_log
group by
  remote_addr
order by
  request_count desc
limit 10;
```

## Example Configurations

### Collect logs from default Apache location

Collect standard Apache access logs from the default location.

```hcl
partition "apache_access_log" "my_apache_logs" {
  source "file" {
    paths       = ["/var/log/apache2/access"]
    file_layout = `%{DATA}.log`
  }
}
```

### Collect logs with custom log format

Define a minimal format that only includes specific fields you need. See the [Apache log configuration documentation](https://httpd.apache.org/docs/current/mod/mod_log_config.html#formats) for a complete list of available format fields.

```hcl
format "apache_access_log" "minimal" {
  layout = `%h %l %u %t` # minimal format with fields client IP, remote logname, remote user, and timestamp
}

partition "apache_access_log" "minimal_logs" {
  source "file" {
    format      = format.apache_access_log.minimal
    paths       = ["/var/log/apache2/minimal"]
    file_layout = `%{DATA}.log`
  }
}
```

### Collect only error responses

Use the filter argument to collect only error responses.

```hcl
partition "apache_access_log" "error_logs" {
  filter = "status >= 400"
  
  source "file" {
    paths       = ["/var/log/apache2/access"]
    file_layout = `%{DATA}.log`
  }
}
```

### Collect logs from multiple virtual hosts

Collect logs from multiple directories or virtual hosts.

```hcl
partition "apache_access_log" "multi_vhost_logs" {
  source "file" {
    paths = [
      "/var/log/apache2/site1/access",
      "/var/log/apache2/site2/access",
      "/var/log/apache2/site3/access"
    ]
    file_layout = `%{DATA}.log`
  }
}
```

### Collect logs from gzip compressed files

If your log files are compressed, you can still collect from them.

```hcl
partition "apache_access_log" "compressed_logs" {
  source "file" {
    paths       = ["/var/log/apache2/archive"]
    file_layout = `%{DATA}.log.gz`
  }
}
```

### Collect logs from ZIP archives

For logs archived in ZIP files, you can collect them directly.

```hcl
partition "apache_access_log" "zip_logs" {
  source "file" {
    paths       = ["/var/log/apache2/archive"]
    file_layout = `%{DATA}.log.zip`
  }
}
```

### Collect logs from S3 bucket

For logs archived in S3, commonly used for long-term storage and centralized logging.

```hcl
connection "aws" "logging" {
  profile = "logging-account"
}

partition "apache_access_log" "s3_logs" {
  source "aws_s3_bucket" {
    connection  = connection.aws.logging
    bucket      = "apache-access-logs"
    prefix      = "logs/"
  }
}
```
