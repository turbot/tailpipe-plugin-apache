---
title: "Tailpipe Table: apache_access_log - Query Apache Access Logs"
description: "Apache access logs capture detailed information about requests processed by the Apache HTTP server. This table provides a structured representation of the log data, including request details, client information, response codes, and processing times."
---

# Table: apache_access_log - Query Apache Access Logs

The `apache_access_log` table allows you to query Apache HTTP server access logs. This table provides detailed information about HTTP requests processed by your Apache servers, including client details, request information, response codes, and timing data.

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

### High Resource Usage Requests

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

### High Traffic Sources

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