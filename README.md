# escope - Elasticsearch CLI Tool

[![Version](https://img.shields.io/github/v/release/mertbahardogan/escope)](https://github.com/mertbahardogan/escope/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/mertbahardogan/escope.svg)](https://pkg.go.dev/github.com/mertbahardogan/escope)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/mertbahardogan/escope)](https://goreportcard.com/report/github.com/mertbahardogan/escope)

**escope** is a powerful CLI tool developed for **diagnostics** and **monitoring** of your **Elasticsearch** cluster. 🚀

## Features

- ⚙️ **Configuration Management** - Save, view, and manage connection settings
- 🔍 **Cluster Health Monitoring** - Quick health status overview with detailed node information
- 📊 **Node Monitoring** - Detailed node metrics and health summary
- 🗑️ **Garbage Collection Analysis** - JVM heap monitoring and GC performance metrics per node
- 📊 **Index Monitoring** - Index health, status, and statistics with alias support, real-time index monitoring with search/index rates and performance metrics
- 🗂️ **Shard Monitoring** - Shard distribution and unassigned shard details (system indices filtered)
- 🔄 **Smart Sorting** - Sort shards and indices by any field with automatic type detection
- 🛡️ **System Index Filtering** - Automatically hides Elasticsearch system indices
- 🔧 **System Information Access** - Dedicated commands for viewing system indices and shards
- 🔬 **Text Analysis** - Analyze text using Elasticsearch analyzers and tokenizers
- ⏱️ **Configurable Timeout** - 3-second timeout for all external API calls
- 🖥️ **TUI Support** - Terminal User Interface with progress bars, colored status badges, and real-time charts

## Requirements

- **Go 1.24.0+** - Required for building and running the application
- **Elasticsearch 7.0.0+** - Compatible with Elasticsearch versions 7.0.0 and above (including 9.0+)

## Installation

```bash
go install github.com/mertbahardogan/escope@latest
```

- After running the installation command, ensure your Go bin directory is included in your system's PATH so you can run `escope` from any location.

## Quick Start

### 1. Set Connection Configuration

```bash
# Save connection settings with alias, multiple alias can be saved
escope config --alias local --host="http://localhost:9200" --username="elastic" --password="password" --secure

# Or for non-secure connections
escope config --alias local --host="http://localhost:9200"
```

### 2. Check Connection

```bash
escope
# Output: Connection successful
```

## Command Reference

| Command | Sub-commands                                                     | Description                                                                           |
|---------|------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| `escope` | `--host`, `--username`, `--password`, `--secure`, `--alias`      | Root command - connection health check and configuration validation                   |
| `escope config` | `list`, `get`, `delete`, `switch`, `current`, `clear`, `timeout` | Multi-host configuration management with alias support and timeout settings           |
| `escope check` | `--duration`, `--interval`                                       | Comprehensive health check across all components with optional continuous monitoring  |
| `escope cluster` | -                                                                | Cluster health overview with node breakdown and shard statistics                      |
| `escope node` | `gc`, `gc --name=<node>`, `dist`                                 | Node health, metrics, garbage collection information, and distribution analysis       |
| `escope index` | `--name=<index>`, `--top`, `system`, `sort`                      | Index status, detailed monitoring, and system indices (filtered by default) |
| `escope shard` | `dist`, `system`, `sort`                                         | Shard analysis, distribution grid, and system shards                                  |
| `escope lucene` | `--name=<index>`                                                 | Lucene segment analysis and memory breakdown (detailed with --name flag)              |
| `escope segments` | -                                                                | Segment count and size analysis per index                                             |
| `escope analyze` | `[analyzer_name] [text] --type`                                  | Analyze text using Elasticsearch analyzer or tokenizer                                |
| `escope termvectors` | `[index] [document_id] [term] --fields`                        | Analyze term vectors and search for specific terms in document fields                 |

## Examples

### Quick Start
```bash
# 1. Set up connection
escope config --alias local --host="http://localhost:9200" --username="elastic" --password="password" --secure

# 2. Test connection
escope
```

### Configuration Management

The tool automatically saves connection settings to local with multi-host alias support:

> **Note:** If you save a configuration with an alias that already exists, it will override the existing configuration. Each alias can only have one configuration at a time.

```bash
# Add a new host with alias
escope config --alias prod --host="http://localhost:9200" --username="elastic" --password="password" --secure

# List all configured hosts
escope config list
# Output:
# Configured hosts:
#   - prod
#   - dev

# View specific host configuration
escope config get prod
# Output:
# Configuration for host 'prod':
#    Host: http://localhost:9200
#    Username: elastic
#    Password: ***
#    Secure: true

# Switch to a different host
escope config switch dev
# Output: Switched to host 'dev'.

# Show currently active host
escope config current
# Output: Active host alias: dev

# Delete a host
escope config delete dev
# Output: Host 'dev' deleted successfully.

# Clear all configurations
escope config clear
# Output: All configurations cleared.

# Timeout Management
# View current timeout setting
escope config timeout
# Output: Current connection timeout: 5 seconds

# Set timeout to 10 seconds
escope config timeout 10
# Output: Connection timeout set to 10 seconds
```

### Cluster Analysis
```bash
# View cluster overview
escope cluster

# Single comprehensive health check
escope check

# Continuous monitoring for 5 minutes
escope check --duration 5m

# High-frequency monitoring (1-second intervals)
escope check --duration 10m --interval 1s

# Check node health and metrics
escope node
```

### Index Monitoring
```bash
# List all indices (system indices filtered)
escope index

# Show system indices
escope index system

# Sort indices by size (largest first)
escope index sort size

# Get single snapshot of index performance metrics
escope index --name my-index
# Output:
#
# Search Rate: -
# Index Rate: -
# Query Time: 15.2 ms
# Index Time: 8.5 ms

# Real-time monitoring (updates every 2 seconds)
escope index --name my-index --top
# Output (refreshes continuously without flicker):
#
# Search Rate: 125.5 /s
# Index Rate: 45.2 /s
# Query Time: 12.8 ms
# Index Time: 22.1 ms
```

### Garbage Collection Monitoring
```bash
# Show GC info for all nodes (sorted by heap usage)
escope node gc
# Output:
# ┌──────────────┬──────────────────┬─────────────────────────┐
# │ Heap Usage % │ Memory Pressure  │ Name                    │
# ├──────────────┼──────────────────┼─────────────────────────┤
# │ 75.3%        │ Medium           │ data-node-1             │
# │ 68.2%        │ Low              │ data-node-2             │
# │ 45.1%        │ Low              │ master-node-1           │
# └──────────────┴──────────────────┴─────────────────────────┘
# Total Nodes: 3
# High Usage (≥80%): 0 (0.0%)
# Medium Usage (60-79%): 2 (66.7%)
# Low Usage (<60%): 1 (33.3%)

# Show detailed GC info for specific node
escope node gc --name=data-node-1

# Analyze node distribution and balance
escope node dist
# Output:
# ┌─────────┬─────────┬───────┬────────┬──────────────┬──────────────────────┐
# │ Primary │ Replica │ Total │ Indices│ IP           │ Name                 │
# ├─────────┼─────────┼───────┼────────┼──────────────┼──────────────────────┤
# │ 15      │ 12      │ 27    │ 8      │ 192.168.1.10 │ elasticsearch-node-1 │
# │ 14      │ 13      │ 27    │ 8      │ 192.168.1.11 │ elasticsearch-node-2 │
# │ 3       │ 0       │ 3     │ 1      │ 192.168.1.12 │ elasticsearch-master │
# └─────────┴─────────┴───────┴────────┴──────────────┴──────────────────────┘
#
# Balance Analysis:
# Most loaded node: elasticsearch-node-1 - 192.168.1.10 (27 shards)
# Least loaded node: elasticsearch-master - 192.168.1.12 (3 shards)
# Balance ratio: 11.1%
# Status: Well balanced
#
# GC Statistics:
#   Young GC:       1250 count / 15.2s total (12.2ms avg)
#   Old GC:         45 count / 8.5s total (188.9ms avg)
#   Full GC:        2 count / 1.2s total (600ms avg)
#
# Performance:
#   GC Frequency:   12.5/min
#   GC Throughput:  98.5%
#   Memory Pressure: Medium
```

### Shard Monitoring
```bash
# View shard status
escope shard

# Show system shards
escope shard system

# View shard distribution across nodes
escope shard dist

# Sort shards by size
escope shard sort size

# Sort shards by state
escope shard sort state
```

### Advanced Analysis
```bash
# Lucene segment analysis (overview of all indices)
escope lucene
# Output:
# ┌──────────┬──────────────┬──────────────┬───────────────┬───────────┬──────────────────────┐
# │ Segments │ Total Memory │ Terms Memory │ Stored Memory │ DocValues │ Index                │
# ├──────────┼──────────────┼──────────────┼───────────────┼───────────┼──────────────────────┤
# │ 10       │ 359.1kb      │ 0b           │ 0b            │ 0b        │ indexName1           │
# │ 2        │ 45.3kb       │ 0b           │ 0b            │ 0b        │ indexName2           │
# └──────────┴──────────────┴──────────────┴───────────────┴───────────┴──────────────────────┘

# Detailed memory breakdown for specific index
escope lucene --name indexName
# Output:
# [Table showing index]
#
# # Index: indexName
#    Segments: 10
#    Total Memory: 359.1kb
#    Index Memory: 
#    Memory Breakdown:
#      • Terms (Inverted Index): 0b
#      • Stored Fields: 0b
#      • DocValues: 0b
#      • Points (Numeric): 0b
#      • Norms: 0b
#      • Fixed BitSet: 359.1kb
#      • Version Map: 0b

# Segment analysis per index
escope segments
# Output:
# ┌──────────┬────────────┬──────────────┬──────────────────────────┐
# │ Segments │ Total Size │ Avg Size/Seg │ Index                    │
# ├──────────┼────────────┼──────────────┼──────────────────────────┤
# │ 24       │ 38mb       │ 1.6mb        │ indexName1               │
# │ 10       │ 373mb      │ 37mb         │ indexName2               │
# └──────────┴────────────┴──────────────┴──────────────────────────┘

# Analyze text using an analyzer
escope analyze standard "Hello World"
# Output:
# +----------+------------+-------+-----+-------+
# | Position | Type       | Start | End | Token |
# +----------+------------+-------+-----+-------+
# | 0        | <ALPHANUM> | 0     | 5   | hello |
# | 1        | <ALPHANUM> | 6     | 11  | world |
# +----------+------------+-------+-----+-------+

# Analyze text with a tokenizer
escope analyze whitespace "Hello World Test" --type tokenizer
# Output:
# +----------+------+-------+-----+-------+
# | Position | Type | Start | End | Token |
# +----------+------+-------+-----+-------+
# | 0        | word | 0     | 5   | Hello |
# | 1        | word | 6     | 11  | World |
# | 2        | word | 12    | 16  | Test  |
# +----------+------+-------+-----+-------+

# Analyze term vectors for a document
escope termvectors my-index doc123 --fields content,title
# Output:
# TERM VECTORS SUMMARY
# ─────────────────────
# Total Terms: 12
# Fields Analyzed: 2
# Highest Frequency: 5
#
# FIELD BREAKDOWN
# ────────────────
#    • content: 8 terms
#    • title: 4 terms
#
# Field: content (8 terms)
# +----------------+-----------+
# | Term           | Frequency |
# +----------------+-----------+
# | elasticsearch  | 5         |
# | search         | 3         |
# | data           | 2         |
# | index          | 2         |
# +----------------+-----------+
#
# Field: title (4 terms)
# +----------------+-----------+
# | Term           | Frequency |
# +----------------+-----------+
# | guide          | 1         |
# | elasticsearch  | 1         |
# +----------------+-----------+

# Search for specific term in document fields
escope termvectors my-index doc123 "elasticsearch" --fields content,title
# Output:
# SEARCH TERM FOUND!
#
# ------- term: elasticsearch -------
#
#  Field            │ Frequency
# ─────────────────────────────────────
#  content          │ 5
#  title            │ 1
```