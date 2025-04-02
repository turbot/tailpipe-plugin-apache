## Activity Examples

### Daily Request Trends

Count requests per day to identify traffic patterns over time. This query provides a comprehensive view of daily request volume, helping you understand usage patterns, peak periods, and potential seasonal variations in web traffic. The results can be used for capacity planning, identifying anomalies, and tracking the impact of site changes or marketing campaigns.

```sql
select
  strftime(timestamp, '%Y-%m-%d') as request_date,
  count(*) as request_count
from
  apache_access_log
group by
  request_date
order by
  request_date asc;
```

### Top 10 Clients by Request Count

List the top 10 client IP addresses making requests. This query helps identify high-volume clients, revealing potential bandwidth abuse, heavy users that might need rate limiting, or unusual access patterns that could indicate automated traffic. Understanding traffic distribution across clients is crucial for optimizing content delivery and resource allocation.

```sql
select
  remote_addr,
  count(*) as request_count,
  count(distinct request_uri) as unique_urls,
  sum(bytes_transferred) as total_bytes
from
  apache_access_log
group by
  remote_addr
order by
  request_count desc
limit 10;
```

### HTTP Status Code Distribution

Analyze the distribution of HTTP status codes across your web traffic. This query provides essential insights into server health and client behavior by breaking down success rates, client errors, and server errors. Understanding this distribution helps identify potential issues, monitor service quality, and track the impact of server configuration changes.

```sql
select
  status,
  count(*) as count,
  round(count(*) * 100.0 / sum(count(*)) over (), 3) as percentage
from
  apache_access_log
group by
  status
order by
  count desc;
```

## Traffic Analysis

### Top HTTP Methods

Analyze the distribution of HTTP methods in your requests. This query reveals how clients interact with your server, helping identify unusual method usage patterns, potential security concerns, and API utilization trends. Understanding method distribution is crucial for security monitoring and ensuring proper server configuration.

```sql
select
  request_method,
  count(*) as request_count,
  round(count(*) * 100.0 / sum(count(*)) over (), 3) as percentage
from
  apache_access_log
group by
  request_method
order by
  request_count desc;
```

### Busiest Days

Identify the days with the highest request volume. This analysis helps optimize resource allocation, plan maintenance windows, and understand traffic patterns across different time periods. The data can be used to correlate traffic spikes with events or promotions and guide infrastructure scaling decisions.

```sql
select
  strftime(timestamp, '%Y-%m-%d') as day,
  count(*) as request_count,
  sum(bytes_transferred) as total_bytes
from
  apache_access_log
group by
  day
order by
  request_count desc;
```

### Busiest Hours

Track hourly traffic patterns to identify peak usage periods. This information is invaluable for scheduling maintenance, optimizing resource allocation, and ensuring adequate capacity during high-traffic periods. Understanding hourly patterns helps in making informed decisions about infrastructure scaling and content delivery optimization.

```sql
select
  date_trunc('hour', timestamp) as hour,
  count(*) as request_count,
  sum(bytes_transferred) as total_bytes
from
  apache_access_log
group by
  hour
order by
  request_count desc;
```

### Most Requested URLs

Analyze the most frequently accessed URLs on your Apache server. This query reveals popular content and high-demand resources, helping optimize caching strategies and content distribution. Understanding URL access patterns is essential for improving user experience and server performance.

```sql
select
  request_uri,
  count(*) as hits,
  avg(bytes_transferred) as avg_bytes
from
  apache_access_log
group by
  request_uri
order by
  hits desc
limit 20;
```

## Error Analysis

### Error Distribution by Status Code

Analyze the distribution of HTTP error responses across your web traffic. This query helps identify specific types of errors affecting your service, their frequency, and potential patterns that might indicate configuration issues or missing resources. Understanding error distribution is crucial for maintaining service quality and user experience.

```sql
select
  status,
  count(*) as error_count
from
  apache_access_log
where
  status >= 400
group by
  status
order by
  error_count desc;
```

## Performance Monitoring

### Large Response Analysis

Identify requests generating substantial data transfer volumes. This query helps detect abnormally large responses that might impact server performance or indicate potential data exfiltration attempts. Understanding large response patterns is crucial for optimizing bandwidth usage, content delivery, and maintaining efficient server operation.

```sql
select
  timestamp,
  remote_addr,
  request_method,
  request_uri,
  bytes_transferred
from
  apache_access_log
where
  bytes_transferred > 1000000  -- More than 1MB
order by
  bytes_transferred desc
limit 20;
```

## User Agent Analysis

### Browser Distribution

Analyze the distribution of client browsers accessing your site. This information helps optimize website compatibility, track mobile versus desktop usage trends, and identify outdated browser versions requiring support. Understanding browser patterns is essential for delivering optimal user experiences across different platforms.

```sql
select
  case
    when http_user_agent like '%Chrome%' then 'Chrome'
    when http_user_agent like '%Firefox%' then 'Firefox'
    when http_user_agent like '%Safari%' then 'Safari'
    when http_user_agent like '%MSIE%' or http_user_agent like '%Trident%' then 'Internet Explorer'
    when http_user_agent like '%Edge%' then 'Edge'
    when http_user_agent like '%bot%' or http_user_agent like '%Bot%' then 'Bot'
    else 'Other'
  end as browser,
  count(*) as request_count
from
  apache_access_log
group by
  browser
order by
  request_count desc;
```

### Bot Traffic Analysis

Monitor and analyze automated traffic patterns across your Apache server. This query helps distinguish between legitimate bot traffic (such as search engine crawlers) and potentially malicious automated access. Understanding bot behavior patterns is crucial for managing server resources and maintaining security.

```sql
select
  http_user_agent,
  count(*) as request_count,
  sum(bytes_transferred) as total_bytes
from
  apache_access_log
where
  regexp_matches(http_user_agent, '(?i)(bot|crawler|spider)')
group by
  http_user_agent
order by
  request_count desc
limit 20;
```

## Security Analysis

### Potential Security Threats

Identify potentially malicious or suspicious requests targeting your server. This query detects common attack patterns, unauthorized access attempts, and potential security vulnerabilities by monitoring request patterns and payload characteristics. Early detection of security threats is essential for maintaining system integrity and protecting sensitive resources.

```sql
select
  timestamp,
  remote_addr,
  request_method,
  request_uri,
  status,
  http_user_agent
from
  apache_access_log
where
  regexp_matches(request_uri, '(?i)(wp-admin|/admin|\.sql|\.git)')
  or request_uri like '%/../%'
  or request_uri like '%<script>%'
  or request_uri like '%union select%'
order by
  timestamp desc
limit 100;
```

### Rate Limiting Analysis

Detect aggressive request patterns that might indicate abuse or denial of service attempts. This query helps identify potential DDoS attacks, aggressive crawlers, or brute force attempts by monitoring request frequency and patterns from individual IP addresses. Understanding request patterns is crucial for implementing effective rate limiting policies.

```sql
select
  remote_addr,
  count(*) as request_count,
  count(distinct request_uri) as unique_urls,
  min(timestamp) as first_request,
  max(timestamp) as last_request
from
  apache_access_log
where
  date_diff('minute', timestamp, cast(current_timestamp as timestamp)) <= 60
group by
  remote_addr
having
  count(*) > 1000
order by
  request_count desc;
```

## Detection Examples

### Failed Authentication Attempts

Track failed authentication attempts across your Apache server. This query helps identify potential brute force attacks or credential stuffing attempts by monitoring patterns of 401 status codes from specific IP addresses. Understanding authentication failure patterns is essential for protecting access to restricted resources and maintaining system security.

```sql
select
  remote_addr,
  count(*) as failed_attempts,
  min(timestamp) as first_attempt,
  max(timestamp) as last_attempt,
  array_agg(distinct request_uri) as attempted_urls
from
  apache_access_log
where
  status = 401
group by
  remote_addr
having
  count(*) > 10
order by
  failed_attempts desc;
```

### Error Spikes

Monitor sudden increases in error rates across your web traffic. This query helps identify potential attacks, system issues, or service degradation by detecting periods where error rates exceed normal thresholds. Understanding error rate patterns is crucial for maintaining service reliability and responding quickly to potential incidents.

```sql
select
  date_trunc('minute', timestamp) as minute,
  count(*) as total_requests,
  count(*) filter (where status >= 400) as error_count,
  round(count(*) filter (where status >= 400) * 100.0 / count(*), 3) as error_rate
from
  apache_access_log
group by
  minute
having
  count(*) > 100 -- Minimum request threshold
and 
  (count(*) filter (where status >= 400) * 100.0 / count(*)) > 20 -- Error rate > 20%
order by
  minute desc;
```

### SQL Injection Attempts

Monitor potential SQL injection attack attempts against your web applications. This query helps identify malicious requests containing SQL syntax patterns that could indicate attempts to manipulate or extract data from backend databases. Understanding SQL injection patterns is crucial for protecting data integrity and preventing unauthorized database access.

```sql
select
  remote_addr,
  request_uri,
  status,
  timestamp,
  http_user_agent
from
  apache_access_log
where
  request_uri like '%SELECT%'
  or request_uri like '%UNION%'
  or request_uri like '%INSERT%'
  or request_uri like '%UPDATE%'
  or request_uri like '%DELETE%'
  or request_uri like '%DROP%'
  or request_uri like '%1=1%'
order by
  timestamp desc;
```

### Geographic Anomalies

Analyze request patterns based on client IP address ranges to identify geographic access anomalies. This query helps detect requests from unusual locations or known problematic regions, aiding in the identification of potential security threats and traffic patterns that may require additional scrutiny or access controls.

```sql
select
  remote_addr,
  count(*) as request_count,
  array_agg(distinct request_uri) as accessed_urls,
  min(timestamp) as first_seen,
  max(timestamp) as last_seen
from
  apache_access_log
where
  remote_addr like '192.%'
  or remote_addr like '10.%'
  or remote_addr like '172.16.%'
group by
  remote_addr
order by
  request_count desc;
```