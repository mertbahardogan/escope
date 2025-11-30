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

	// Replica thresholds
	OptimalReplicaCount       = 2 // Optimal replica count
	MaxAcceptableReplicaCount = 3 // Maximum acceptable without warning

	// Rate-based thresholds for dynamic shard recommendations
	LowRateThreshold      = 10.0   // requests/sec - low traffic
	MediumRateThreshold   = 100.0  // requests/sec - medium traffic
	HighRateThreshold     = 1000.0 // requests/sec - high traffic
	VeryHighRateThreshold = 5000.0 // requests/sec - very high traffic

	// Document count based thresholds
	MinDocsPerShard     = 1000000  // 1M docs minimum per shard
	OptimalDocsPerShard = 10000000 // 10M docs optimal per shard
	MaxDocsPerShard     = 50000000 // 50M docs maximum per shard

	// Scoring weights for hybrid recommendation (must sum to 1.0)
	SizeWeight     = 0.5 // Weight for size-based calculation
	TrafficWeight  = 0.3 // Weight for traffic-based calculation
	DocCountWeight = 0.2 // Weight for document count calculation

	// Acceptable range flexibility (as percentage)
	AcceptableRangeFlexibility = 0.4 // ±40% from recommended is acceptable

	// Severity thresholds (deviation percentage from recommended)
	SeverityCriticalThreshold = 100.0 // >100% deviation
	SeverityHighThreshold     = 60.0  // 60-100% deviation
	SeverityMediumThreshold   = 40.0  // 40-60% deviation
	SeverityLowThreshold      = 30.0  // 30-40% deviation

	// Confidence thresholds
	HighConfidence   = 0.8 // High confidence in recommendation
	MediumConfidence = 0.5 // Medium confidence

	// Minimum index size to check (ignore very small indices)
	MinIndexSizeForCheck = 1 * BytesInGB // 1GB minimum

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

	GCYoung                     = "young"
	GCOld                       = "old"
	GCSurvivor                  = "survivor"
	GCG1Concurrent              = "G1 Concurrent GC"
	GCField                     = "gc"
	CollectorsField             = "collectors"
	CollectionCountField        = "collection_count"
	CollectionTimeInMillisField = "collection_time_in_millis"
	PoolsField                  = "pools"

	MemoryPressureLow    = "Low"
	MemoryPressureMedium = "Medium"
	MemoryPressureHigh   = "High"

	PercentFormat    = "%.0f%%"
	RateFormatK      = "%.1f K/s"
	RateFormat       = "%.1f /s"
	RateFormat2      = "%.2f /s"
	TimeFormatMS     = "%.1f ms"
	TimeFormatS      = "%.1f s"
	MSFormat         = "%d ms"
	GCFreqFormat     = "%.1f GC/sec"
	ThroughputFormat = "%.1f%%"

	ByteSuffix = "b"
	KiloSuffix = "kb"
	MegaSuffix = "mb"
	GigaSuffix = "gb"
	TeraSuffix = "tb"

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

	ScaleStateOverScaled      = "OVER-SCALED"
	ScaleStateUnderScaled     = "UNDER-SCALED"
	ScaleStateOverReplicated  = "OVER-REPLICATED"
	ScaleStateUnderReplicated = "UNDER-REPLICATED"

	WarningTypeOverScaled      = "over-scaled"
	WarningTypeUnderScaled     = "under-scaled"
	WarningTypeOverReplicated  = "over-replicated"
	WarningTypeUnderReplicated = "under-replicated"

	RecommendationCategoryShard   = "SHARD"
	RecommendationCategoryIndex   = "INDEX"
	RecommendationCategoryNode    = "NODE"
	RecommendationCategoryGeneral = "GENERAL"
)
