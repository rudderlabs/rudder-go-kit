package clickhouse

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
)

const clusterName = "rudder_cluster"

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
