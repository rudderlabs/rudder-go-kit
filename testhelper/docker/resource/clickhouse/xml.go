package clickhouse

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
)

const clusterName = "rudder_cluster"

// testTuningSettings sizes the server for short-lived tests. The README gives the measured savings. The values
// go into resource.xml, which sorts before the caller's zz_extra.xml, so WithConfig still overrides any value.
//
// Constraints found by measurement and in the ClickHouse source:
//   - background_pool_size × background_merges_mutations_concurrency_ratio must be at least 25, the default of
//     number_of_free_entries_in_pool_to_execute_optimize_entire_partition. Below that the server exits at startup
//     while the container stays up. The ratio sizes only the task queue, not the threads, so 10 keeps room for
//     MergeTree settings that stock accepts.
//   - background_fetches_pool_size stays at 4: the fetch limit is server-wide, so a smaller pool makes replicas
//     lag behind inserts in cluster mode.
//   - max_thread_pool_size is not set: a cap below the server's idle thread need hangs startup, and threads are
//     created on demand anyway.
//   - max_server_memory_usage is not set: ClickHouse derives it from the container limit, and a fixed value
//     trips during startup or breaks WithMemory(0).
//   - query_log stays enabled, because tests read it after SYSTEM FLUSH LOGS.
//   - background_distributed_schedule_pool_size is not set: a pool of 1 logs "Temporarily pause scheduling of
//     tasks" on every distributed send of async inserts.
var testTuningSettings = []struct{ name, value string }{
	{"background_pool_size", "5"},
	{"background_merges_mutations_concurrency_ratio", "10"},
	{"background_schedule_pool_size", "4"},
	{"background_move_pool_size", "1"},
	{"background_fetches_pool_size", "4"},
	{"background_common_pool_size", "2"},
	{"mark_cache_size", "16777216"},
	{"uncompressed_cache_size", "0"},
	{"index_mark_cache_size", "0"},
	{"asynchronous_metrics_update_period_s", "600"},
}

// testTuningRemovedLogs are the system log tables the test tuning disables. Removing an element that the image
// does not define has no effect, so older tags stay safe.
var testTuningRemovedLogs = []string{
	"metric_log", "trace_log", "text_log", "asynchronous_metric_log", "asynchronous_insert_log", "part_log",
	"processors_profile_log", "opentelemetry_span_log", "query_thread_log", "query_views_log", "crash_log",
	"background_schedule_pool_log",
}

func testTuningXML() string {
	var document strings.Builder
	for _, setting := range testTuningSettings {
		document.WriteString("<" + setting.name + ">" + setting.value + "</" + setting.name + ">")
	}
	for _, table := range testTuningRemovedLogs {
		document.WriteString("<" + table + ` remove="1"/>`)
	}
	return document.String()
}

type Macros struct {
	Cluster string `xml:"cluster"`
	Shard   string `xml:"shard"`
	Replica string `xml:"replica"`
}

func topology(prefix string, config Config) []*Node {
	shards, replicas := config.Shards, config.Replicas
	if shards == 0 {
		shards, replicas = 1, 1
	}
	nodes := make([]*Node, 0, shards*replicas)
	for shard := 1; shard <= shards; shard++ {
		for replica := 1; replica <= replicas; replica++ {
			hostname := prefix + "-s" + strconv.Itoa(shard) + "-r" + strconv.Itoa(replica)
			node := &Node{Hostname: hostname, User: config.User, Password: config.Password, Database: config.Database, secure: config.TLS}
			if config.Shards > 0 {
				node.Macros = Macros{Cluster: clusterName, Shard: strconv.Itoa(shard), Replica: hostname}
			}
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func escapeXML(value string) string {
	var buffer bytes.Buffer
	if err := xml.EscapeText(&buffer, []byte(value)); err != nil {
		panic(err)
	}
	return buffer.String()
}

func serverConfig(config Config, nodes []*Node, index int) string {
	var document strings.Builder
	document.WriteString(`<clickhouse><logger><level>warning</level><console>1</console></logger>`)
	if !config.DisableTestTuning {
		document.WriteString(testTuningXML())
	}
	if config.TLS {
		document.WriteString(`<https_port>8443</https_port><tcp_port_secure>9440</tcp_port_secure><openSSL><server><certificateFile>/etc/clickhouse-server/certs/server.pem</certificateFile><privateKeyFile>/etc/clickhouse-server/certs/server.key</privateKeyFile><verificationMode>none</verificationMode><loadDefaultCAFile>false</loadDefaultCAFile><disableProtocols>sslv2,sslv3,tlsv1,tlsv1_1</disableProtocols></server></openSSL>`)
	}
	if config.Shards > 0 {
		document.WriteString(`<remote_servers><` + clusterName + `>`)
		for shard := 0; shard < config.Shards; shard++ {
			document.WriteString(`<shard><internal_replication>true</internal_replication>`)
			for replica := 0; replica < config.Replicas; replica++ {
				node := nodes[shard*config.Replicas+replica]
				document.WriteString(`<replica><host>` + node.Hostname + `</host><port>9000</port><user>` + escapeXML(config.User) + `</user><password>` + escapeXML(config.Password) + `</password></replica>`)
			}
			document.WriteString(`</shard>`)
		}
		document.WriteString(`</` + clusterName + `></remote_servers>`)
		document.WriteString(`<zookeeper><node><host>` + nodes[0].Hostname + `</host><port>9181</port></node></zookeeper>`)
		document.WriteString(`<distributed_ddl><path>/clickhouse/task_queue/ddl</path></distributed_ddl>`)
		node := nodes[index]
		document.WriteString(`<interserver_http_host>` + node.Hostname + `</interserver_http_host><interserver_http_port>9009</interserver_http_port>`)
		document.WriteString(`<macros><cluster>` + clusterName + `</cluster><shard>` + node.Macros.Shard + `</shard><replica>` + node.Macros.Replica + `</replica></macros>`)
		if index == 0 {
			document.WriteString(`<keeper_server><tcp_port>9181</tcp_port><server_id>1</server_id><log_storage_path>/var/lib/clickhouse/coordination/log</log_storage_path><snapshot_storage_path>/var/lib/clickhouse/coordination/snapshots</snapshot_storage_path><raft_configuration><server><id>1</id><hostname>` + node.Hostname + `</hostname><port>9234</port></server></raft_configuration></keeper_server>`)
		}
	}
	document.WriteString(`</clickhouse>`)
	return document.String()
}
