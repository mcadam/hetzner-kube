package cmd

import (
	"testing"
	"github.com/andreyvit/diff"
)

func TestGenerateMasterConfiguration(t *testing.T) {
	expectedConf := `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
api:
  advertiseAddress: 10.0.0.1
networking:
  podSubnet: 10.244.0.0/16
apiServerCertSANs:
  - 1.1.1.1
  - 127.0.0.1
  - 10.0.0.1
  - 10.0.0.2
`

	expectedConfWithEtcd := `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
api:
  advertiseAddress: 10.0.0.1
networking:
  podSubnet: 10.244.0.0/16
apiServerCertSANs:
  - 1.1.1.1
  - 127.0.0.1
  - 10.0.0.1
  - 10.0.0.2
etcd:
  endpoints:
  - http://10.0.0.1:2379
  - http://10.0.0.2:2379
`
	nodes := []Node{
		{Name: "node1", IPAddress: "1.1.1.1", PrivateIPAddress: "10.0.0.1", },
		{Name: "node2", IPAddress: "1.1.1.2", PrivateIPAddress: "10.0.0.2", },
	}

	noEtcdConf := GenerateMasterConfiguration(nodes[0], nodes,nil)

	if noEtcdConf != expectedConf {
		t.Errorf("master config without etcd does not match to expected.\n%s\n", noEtcdConf)
	}

	etcdConf := GenerateMasterConfiguration(nodes[0], nodes, nodes)

	if etcdConf != expectedConfWithEtcd {
		t.Errorf("master config with etcd does not match to expected.\n%s\n", diff.LineDiff(etcdConf, expectedConfWithEtcd))
	}

}

func TestGenerateEtcdSystemdService(t *testing.T) {
	expectedString := `# /etc/systemd/system/etcd.service
[Unit]
Description=etcd
After=network.target wg-quick@wg0.service

[Service]
ExecStart=/opt/etcd/etcd --name kube1 \
  --data-dir /var/lib/etcd \
  --listen-client-urls "http://10.0.1.11:2379,http://localhost:2379" \
  --advertise-client-urls "http://10.0.1.11:2379" \
  --listen-peer-urls "http://10.0.1.11:2380" \
  --initial-cluster "kube1=http://10.0.1.11:2380,kube2=http://10.0.1.12:2380,kube3=http://10.0.1.13:2380" \
  --initial-advertise-peer-urls "http://10.0.1.11:2380" \
  --heartbeat-interval 200 \
  --election-timeout 5000
Restart=always
RestartSec=5
TimeoutStartSec=0
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
`
	nodes := []Node{
		{Name: "kube1", IPAddress: "1.1.1.1", PrivateIPAddress: "10.0.1.11", },
		{Name: "kube2", IPAddress: "1.1.1.2", PrivateIPAddress: "10.0.1.12", },
		{Name: "kube3", IPAddress: "1.1.1.3", PrivateIPAddress: "10.0.1.13", },
	}

	etcdService := GenerateEtcdSystemdService(nodes[0], nodes)

	if etcdService != expectedString {
		t.Errorf("etcd systemd service does not match expected\n%s", diff.LineDiff(expectedString, etcdService))
	}

}

func TestCluster_CreateEtcdNodes(t *testing.T) {
	nodes := []Node{
		{Name: "kube1", IPAddress: "1.1.1.1", PrivateIPAddress: "10.0.1.11", IsEtcd: true },
		{Name: "kube2", IPAddress: "1.1.1.2", PrivateIPAddress: "10.0.1.12", IsMaster: true},
		{Name: "kube3", IPAddress: "1.1.1.3", PrivateIPAddress: "10.0.1.13", },
	}

	cluster := Cluster{Nodes:nodes}

	etcdNodes := cluster.GetEtcdNodes()

	if len(etcdNodes) != 1 {
		t.Error("found more than one etcd node")
	}

	if etcdNodes[0].Name != nodes[0].Name {
		t.Error("wrong node found")
	}
}

