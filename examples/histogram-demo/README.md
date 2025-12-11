# Exponential Histogram Demo

This demo showcases the exponential histogram feature in rudder-go-kit with Prometheus and Grafana visualization.

## Overview

The demo generates simulated request latencies ranging from 1ms to 5 seconds and exports them as exponential histograms 
to Prometheus. 
Exponential histograms provide better accuracy and lower memory usage for high-dynamic-range metrics compared to 
traditional fixed-bucket histograms.

## Prerequisites

- Docker and Docker Compose
- Go 1.25.4 or later

## Quick Start

### 1. Start the Infrastructure

Navigate to the demo directory and start Prometheus and Grafana containers:

```bash
cd examples/histogram-demo
docker-compose up -d
```

This will start:
- **Prometheus** on http://localhost:9090
- **Grafana** on http://localhost:3000 (credentials: admin/admin)

### 2. Run the Histogram Generator

```bash
cd examples/histogram-demo/app
go run main.go -duration 5
```

Options:
- `-duration`: Number of minutes to generate data (default: 5)
- `-port`: Port for Prometheus metrics endpoint (default: 8080)

The script will:
- Generate random latencies from 1ms to 5s with a realistic distribution:
  - 50% between 1-50ms (fast requests)
  - 25% between 50-500ms (normal requests)
  - 15% between 500ms-2s (slow requests)
  - 7% between 2-4s (very slow requests)
  - 3% between 4-5s (timeout-prone requests)
- Export metrics on http://localhost:8080/metrics
- Run for the specified duration
- Keep containers running after completion for analysis

### 3. Explore the Data

#### Prometheus

Visit http://localhost:9090 and query:
- `request_latency_seconds` - View the native histogram data
- `histogram_quantile(0.95, request_latency_seconds)` - Calculate P95 latency
- `histogram_quantile(0.99, request_latency_seconds)` - Calculate P99 latency

#### Grafana

1. Visit http://localhost:3000 (login: admin/admin)

2. **Configure Prometheus Data Source** (if not already configured):
   - Click on the hamburger menu (☰) in the top left
   - Go to **Connections** → **Data sources**
   - Click **Add data source**
   - Select **Prometheus**
   - Configure with:
     - **Name**: `Prometheus`
     - **URL**: `http://prometheus:9090`
     - **HTTP Method**: `POST` (recommended for native histograms)
   - Click **Save & Test** - you should see "Successfully queried the Prometheus API"

3. Create a new dashboard and add panels with queries like:
   - Histogram visualization: `request_latency_seconds`
   - Heatmap: `rate(request_latency_seconds[5m])`
   - Percentiles: `histogram_quantile(0.95, request_latency_seconds)`

**Tip**: For native histogram visualization in Grafana, use the **Heatmap** panel type with the query 
`rate(request_latency_seconds[5m])` to see the distribution over time.

## Understanding Native Histograms

Native histograms (exponential histograms) provide:

1. **Better Accuracy**: Automatically adapt to data distribution
2. **Lower Memory**: Use fewer buckets with exponential spacing
3. **No Configuration**: No need to pre-define bucket boundaries
4. **Efficient Storage**: Sparse representation saves space

The `maxSize` parameter (160 in this demo) controls the maximum number of buckets, providing a good balance between 
accuracy and resource usage.

## Cleanup

Stop and remove containers:

```bash
docker-compose down -v
```

## How It Works

The demo uses rudder-go-kit's stats package with:

```go
stats.NewStats(conf, log, m,
    stats.WithDefaultExponentialHistogram(160),
)
```

This configures all histograms to use exponential bucketing with up to 160 buckets. The data is then exported to 
Prometheus in the native histogram format, which Prometheus can efficiently store and query.

## Prometheus Configuration Note

The Prometheus configuration includes `--enable-feature=native-histograms` which is required to handle exponential 
histograms correctly.
