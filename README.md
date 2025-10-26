# escope - Elasticsearch CLI Tool

**escope** is a powerful CLI tool developed for **diagnostics** and **monitoring** of your **Elasticsearch** cluster. 🚀
## Features

- ⚙️ **Configuration Management** - Save, view, and manage connection settings
- 🔍 **Cluster Health Monitoring** - Quick health status overview with detailed node information
- 📊 **Node Information** - Detailed node metrics and health summary
- 🗑️ **Garbage Collection Analysis** - JVM heap monitoring and GC performance metrics per node
- 📁 **Index Management** - Index health, status, and statistics with alias support (system indices filtered)
- 📊 **Index Monitoring** - Real-time index monitoring with search/index rates and performance metrics
- 🗂️ **Shard Analysis** - Shard distribution and unassigned shard details (system indices filtered)
- 🔄 **Smart Sorting** - Sort shards and indices by any field with automatic type detection
- 🛡️ **System Index Filtering** - Automatically hides Elasticsearch system indices
- 🔧 **System Information Access** - Dedicated commands for viewing system indices and shards
- 🔬 **Text Analysis** - Analyze text using Elasticsearch analyzers and tokenizers
- ⏱️ **Configurable Timeout** - 3-second timeout for all external API calls

## Requirements

- **Go 1.24.0+** - Required for building and running the application
- **Elasticsearch 7.0.0+** - Compatible with Elasticsearch versions 7.0.0 and above (including 9.0+)
- **Network Access** - Access to your Elasticsearch cluster endpoints
- **Authentication** - Valid credentials for your Elasticsearch cluster (if authentication is enabled)

## Installation

```bash
go install github.com/mertbahardogan/escope@latest
```

After running the installation command, ensure your Go bin directory is included in your system's PATH so you can run `escope` from any location.

Once installed, **escope** is ready to use. If no configuration exists, the tool will provide helpful setup instructions when you first run it.

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

| Command | Sub-commands                                                     | Description |
|---------|------------------------------------------------------------------|-------------|
| `escope` | `--host`, `--username`, `--password`, `--secure`, `--alias`      | Root command - connection health check and configuration validation |
| `escope config` | `list`, `get`, `delete`, `switch`, `current`, `clear`, `timeout` | Multi-host configuration management with alias support and timeout settings |
| `escope check` | `--duration`, `--interval`                                       | Comprehensive health check across all components with optional continuous monitoring |
| `escope cluster` | -                                                                | Cluster health overview with node breakdown and shard statistics |
| `escope node` | `gc`, `gc --name=<node>`, `dist`                                 | Node health, metrics, garbage collection information, and distribution analysis |
| `escope index` | `--name=<index>`, `--top`, `system`, `sort`                      | Index management, status, detailed monitoring, and system indices (filtered by default) |
| `escope shard` | `dist`, `system`, `sort`                                         | Shard analysis, distribution grid, and system shards |
| `escope lucene` | `--name=<index>`                                                 | Lucene segment analysis and memory breakdown (detailed with --name flag) |
| `escope segments` | -                                                                | Segment count and size analysis per index |
| `escope analyze` | `[analyzer_name] [text] --type`                                  | Analyze text using Elasticsearch analyzer or tokenizer |
| `escope termvectors` | `[index] [document_id] [term] --fields`                        | Analyze term vectors and search for specific terms in document fields |

## Configuration

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
# Output: Switched to host 'dev'. All commands will now use this host.

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

## Examples

### Quick Start
```bash
# 1. Set up connection
escope config --alias local --host="http://localhost:9200" --username="elastic" --password="password" --secure

# 2. Test connection
escope

# 3. Check cluster health
escope cluster
```

### Health Monitoring
```bash
# Single comprehensive health check
escope check

# Continuous monitoring for 5 minutes
escope check --duration 5m

# High-frequency monitoring (1-second intervals)
escope check --duration 10m --interval 1s
```

### Cluster Analysis
```bash
# View cluster overview
escope cluster

# Check node health and metrics
escope node

# Show garbage collection info for all nodes
escope node gc

# Show detailed GC info for specific node
escope node gc --name=data-node-1

# Analyze shard distribution
escope shard

# View shard distribution across nodes
escope shard dist
```

### Index Management
```bash
# List all indices (system indices filtered)
escope index

# Show detailed information for a specific index
escope index --name my-index

# Monitor index in real-time (like top command)
escope index --name my-index --top

# Show system indices
escope index system

# Sort indices by size (largest first)
escope index sort size

# Sort indices by document count
escope index sort docs
```

### Index Monitoring
```bash
# Get single snapshot of index performance metrics
escope index --name my-index
# Output:
# ┌─────────────────────────────────────────────────┐
# │ my-index                                        │
# ├─────────────────────────────────────────────────┤
# │ Search Rate: Calculating...                     │
# │ Index Rate: Calculating...                      │
# │ Query Time: 15.2ms                              │
# │ Index Time: 8.5ms                               │
# └─────────────────────────────────────────────────┘

# Real-time monitoring (updates every 2 seconds)
escope index --name my-index --top
# Output (refreshes continuously):
# ┌─────────────────────────────────────────────────┐
# │ my-index | Check 5                              │
# ├─────────────────────────────────────────────────┤
# │ Search Rate: 125.5/s                            │
# │ Index Rate: 45.2/s                              │
# │ Query Time: 12.8ms                              │
# │ Index Time: 22.1ms                              │
# └─────────────────────────────────────────────────┘
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

### Shard Analysis
```bash
# View shard status
escope shard

# Show system shards
escope shard system

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
# [Shows term vector analysis for the document]

# Search for specific term in document fields
escope termvectors my-index doc123 "term" --fields content,title
# Output:
# [Shows search results for the specific term in the document]
```

### Configuration Management
```bash
# Add multiple hosts
escope config --alias prod --host="http://localhost:9200" --username="admin" --password="secret" --secure
escope config --alias dev --host="http://localhost:9200"

# List all configurations
escope config list

# Switch between hosts
escope config switch prod

# View current configuration
escope config current

# Manage timeout settings
escope config timeout          # View current timeout
escope config timeout 10       # Set timeout to 10 seconds
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for detailed information on how to contribute to this project.

### Quick Start for Contributors

1. **Fork the repository** on GitHub
2. **Create a feature branch**: `git checkout -b feat/your-feature-name`
3. **Make your changes** and test them
4. **Run the test suite**: `make test-commands` (mandatory)
5. **Submit a pull request**

### Development Workflow

- Always create feature branches (`feat/`, `fix/`, `docs/`)
- Run `make test-commands` before submitting any changes
- Follow our coding standards and commit message format
- Update documentation for new features

For more details, see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- Check existing issues for solutions
- Create an issue on GitHub
- Review the documentation