---
organization: Turbot
category: ["software development"]
icon_url: "/images/plugins/turbot/apache.svg"
brand_color: "#CC2336"
display_name: "Apache"
description: "Tailpipe plugin for collecting and querying Apache logs."
og_description: "Collect Apache logs and query them instantly with SQL! Open source CLI. No DB required."
og_image: "/images/plugins/turbot/apache-social-graphic.png"
---

# Apache + Tailpipe

[Tailpipe](https://tailpipe.io) is an open-source CLI tool that allows you to collect logs and query them with SQL.

The [Apache Plugin for Tailpipe](https://hub.tailpipe.io/plugins/turbot/apache) allows you to collect and query Apache access logs using SQL to track activity, monitor trends, detect anomalies, and more!

- **[Get started →](https://hub.tailpipe.io/plugins/turbot/apache)**
- Documentation: [Table definitions & examples](https://hub.tailpipe.io/plugins/turbot/apache/tables)
- Community: [Join #tailpipe on Slack →](https://turbot.com/community/join)
- Get involved: [Issues](https://github.com/turbot/tailpipe-plugin-apache/issues)

![image](https://raw.githubusercontent.com/turbot/tailpipe-plugin-apache/main/docs/images/apache_access_log_terminal.png?type=thumbnail)

![image](https://raw.githubusercontent.com/turbot/tailpipe-plugin-apache/main/docs/images/apache_access_log_owasp_top_10_dashboard.png?type=thumbnail)

## Getting Started

Install Tailpipe from the [downloads](https://tailpipe.io/downloads) page:

```sh
# MacOS
brew install turbot/tap/tailpipe
```

```sh
# Linux or Windows (WSL)
sudo /bin/sh -c "$(curl -fsSL https://tailpipe.io/install/tailpipe.sh)"
```

Install the plugin:

```sh
tailpipe plugin install apache
```

Configure your table partition and data source:

```sh
vi ~/.tailpipe/config/apache.tpc
```

```hcl
partition "apache_access_log" "my_logs" {
  source "file" {
    paths       = ["/var/log/apache/access/"]
    file_layout = `%{DATA}.log`
  }
}
```

**Note**: By default, the `apache_access_log` table can collect logs using the [common](https://httpd.apache.org/docs/current/logs.html#common) and [combined](https://httpd.apache.org/docs/current/logs.html#combined) log formats. If your logs use a custom log format, please see [Collect logs with custom log format](https://hub.tailpipe.io/plugins/turbot/apache/tables/apache_access_log#collect-logs-with-custom-log-format).

Download, enrich, and save logs from your source ([examples](https://tailpipe.io/docs/reference/cli/collect)):

```sh
tailpipe collect apache_access_log
```

Enter interactive query mode:

```sh
tailpipe query
```

Run a query:

```sql
select
  remote_addr,
  status,
  request_uri,
  request_method,
  count(*) as request_count
from 
  apache_access_log
group by 
  remote_addr, 
  status, 
  request_uri, 
  request_method
order by 
  request_count desc
limit 1;
```

```sh
+-----------------+--------+-------------------+----------------+---------------+
| remote_addr     | status | request_uri       | request_method | request_count |
+-----------------+--------+-------------------+----------------+---------------+
| 186.187.161.169 | 502    | /path/to/web/page | POST           | 12345         |
+-----------------+--------+-------------------+----------------+---------------+
```

## Detections as Code with Powerpipe

Pre-built dashboards and detections for the Apache plugin are available in [Powerpipe](https://powerpipe.io) mods, helping you monitor and analyze activity across your Apache servers.

For example, the [Apache Access Log Detections mod](https://hub.powerpipe.io/mods/turbot/tailpipe-mod-apache-access-log-detections) scans your Apache access logs for anomalies, such as sql injestion attacks on your web application.

Dashboards and detections are [open source](https://github.com/topics/tailpipe-mod), allowing easy customization and collaboration.

To get started, choose a mod from the [Powerpipe Hub](https://hub.powerpipe.io/?engines=tailpipe&q=apache).
