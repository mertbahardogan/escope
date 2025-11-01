package constants

const (
	DefaultInterval = 2
	MinInterval     = 1

	HealthGreen  = "green"
	HealthYellow = "yellow"
	HealthRed    = "red"

	SeverityCritical = "CRITICAL"
	SeverityWarning  = "WARNING"

	HealthField    = "health"
	StatusField    = "status"
	IndexField     = "index"
	DocsCountField = "docs.count"
	StoreSizeField = "store.size"
	PrimaryField   = "pri"
	ReplicaField   = "rep"

	HeapUsedPctField = "heap_used_percent"
	CPUPercentField  = "percent"

	ShardField = "shard"
	StateField = "state"
	StoreField = "store"

	ClusterNameField             = "cluster_name"
	NumberOfNodesField           = "number_of_nodes"
	ActivePrimaryShardsField     = "active_primary_shards"
	ActiveShardsField            = "active_shards"
	UnassignedShardsField        = "unassigned_shards"
	RelocatingShardsField        = "relocating_shards"
	InitializingShardsField      = "initializing_shards"
	DelayedUnassignedShardsField = "delayed_unassigned_shards"
	NumberOfPendingTasksField    = "number_of_pending_tasks"
	NumberOfInFlightFetchField   = "number_of_in_flight_fetch"
	TaskMaxWaitingInQueueField   = "task_max_waiting_in_queue_millis"
	ActiveShardsPercentField     = "active_shards_percent_as_number"
	TimedOutField                = "timed_out"

	NodesField            = "nodes"
	OSField               = "os"
	CPUField              = "cpu"
	JVMMemField           = "mem"
	FSField               = "fs"
	TotalField            = "total"
	TotalInBytesField     = "total_in_bytes"
	AvailableInBytesField = "available_in_bytes"

	IndicesField           = "indices"
	IndexingField          = "indexing"
	SearchField            = "search"
	IndexTotalField        = "index_total"
	IndexTimeInMillisField = "index_time_in_millis"
	QueryTotalField        = "query_total"
	QueryTimeInMillisField = "query_time_in_millis"

	ShardStateStarted      = "STARTED"
	ShardStateInitializing = "INITIALIZING"
	ShardStateRelocating   = "RELOCATING"
	ShardStateUnassigned   = "UNASSIGNED"

	NodeRoleData   = "data"
	NodeRoleMaster = "master"
	NodeRoleIngest = "ingest"

	DefaultTimeout = 5

	// Elasticsearch field keys
	CountField                     = "count"
	MemoryInBytesField             = "memory_in_bytes"
	TermsMemoryInBytesField        = "terms_memory_in_bytes"
	TermsField                     = "terms"
	StoredFieldsMemoryInBytesField = "stored_fields_memory_in_bytes"
	StoredFieldsField              = "stored_fields"
	DocValuesMemoryInBytesField    = "doc_values_memory_in_bytes"
	DocValuesField                 = "doc_values"
	PointsMemoryInBytesField       = "points_memory_in_bytes"
	PointsField                    = "points"
	NormsMemoryInBytesField        = "norms_memory_in_bytes"
	NormsField                     = "norms"
	FixedBitSetMemoryInBytesField  = "fixed_bit_set_memory_in_bytes"
	VersionMapMemoryInBytesField   = "version_map_memory_in_bytes"
	MaxUnsafeAutoIDTimestampField  = "max_unsafe_auto_id_timestamp"
	IndexMemoryField               = "index_memory"
	SegmentsField                  = "segments"

	// Node field keys
	NameField            = "name"
	IPField              = "ip"
	RolesField           = "roles"
	ProcessField         = "process"
	JVMField             = "jvm"
	MemField             = "mem"
	HeapUsedInBytesField = "heap_used_in_bytes"
	HeapMaxInBytesField  = "heap_max_in_bytes"
	HeapUsedPercentField = "heap_used_percent"
	UsedInBytesField     = "used_in_bytes"
	UsedPercentField     = "used_percent"
	PercentField         = "percent"
	DocsField            = "docs"
	StoreFieldKey        = "store"

	// Shard field keys
	NodeFieldKey = "node"
	IPFieldKey   = "ip"
	AliasField   = "alias"
	PrirepField2 = "prirep"

	// String values
	EmptyString        = ""
	DashString         = "-"
	PrimaryShortString = "p"
	ReplicaShortString = "r"
	ZeroByteString     = "0b"
	HealthyString      = "healthy"
	WarningString      = "warning"
	PrimaryString      = "Primary"
	ReplicaString      = "Replica"

	// Numeric thresholds
	HighSegmentThreshold  = 50
	SmallSegmentThreshold = 1024 * 1024        // 1MB
	LargeSegmentThreshold = 1024 * 1024 * 1024 // 1GB
	HighCPUThreshold      = 80
	HighMemoryThreshold   = 90
	HighHeapThreshold     = 85
	HighDiskThreshold     = 90
	BalanceRatioThreshold = 0.7
	LowMemoryPressure     = 60
	MediumMemoryPressure  = 80

	// Byte conversion constants
	BytesInKB = 1024
	BytesInMB = 1024 * 1024
	BytesInGB = 1024 * 1024 * 1024
	BytesInTB = 1024 * 1024 * 1024 * 1024

	// Config defaults
	DefaultConfigTimeout  = 3
	DefaultConfigTimeout2 = 30
	ConfigFilePath        = ".escope.yaml"
	ConfigFileEnvPath     = "$HOME/.escope.yaml"

	// GC collector names
	GCYoung        = "young"
	GCOld          = "old"
	GCSurvivor     = "survivor"
	GCG1Concurrent = "G1 Concurrent GC"

	// GC fields
	GCField                     = "gc"
	CollectorsField             = "collectors"
	CollectionCountField        = "collection_count"
	CollectionTimeInMillisField = "collection_time_in_millis"
	PoolsField                  = "pools"

	// Memory pressure levels
	MemoryPressureLow    = "Low"
	MemoryPressureMedium = "Medium"
	MemoryPressureHigh   = "High"

	// Format strings
	PercentFormat    = "%.0f%%"
	RateFormatK      = "%.1f K/s"
	RateFormat       = "%.1f /s"
	RateFormat2      = "%.2f /s"
	TimeFormatMS     = "%.1f ms"
	TimeFormatS      = "%.1f s"
	MSFormat         = "%d ms"
	GCFreqFormat     = "%.1f GC/sec"
	ThroughputFormat = "%.1f%%"

	// Size unit suffixes
	ByteSuffix = "b"
	KiloSuffix = "kb"
	MegaSuffix = "mb"
	GigaSuffix = "gb"
	TeraSuffix = "tb"

	// System index prefixes
	DotPrefix        = "."
	KibanaPrefix     = "kibana"
	APMPrefix        = "apm"
	SecurityPrefix   = "security"
	MonitoringPrefix = "monitoring"
	WatcherPrefix    = "watcher"
	ILMPrefix        = "ilm"
	SLMPrefix        = "slm"
	TransformPrefix  = "transform"

	ThousandDivisor       = 1000
	HundredMultiplier     = 100
	DocsCountSeparator    = 3
	TenThreshold          = 10
	ZeroPercentString     = "0%"
	MillisecondsToSeconds = 1000

	ANSIClearScreen  = "\033[2J\033[H" // Clear screen and move cursor to home
	ANSIClearLineEnd = "\033[K"        // Clear from cursor to end of line
	ANSIMoveUpFormat = "\033[%dA\r"    // Move cursor up N lines and return to start

	FirstCheckCount      = 1
	IndexDetailLineCount = 5
)
