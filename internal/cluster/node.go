package cluster

import "time"

type NodeStatus string

const (
	StatusAlive      NodeStatus = "alive"
	StatusDead       NodeStatus = "dead"
	StatusRecovering NodeStatus = "recovering"
)

type NodeInfo struct {
	ID        string     `json:"id"`
	HTTPAddr  string     `json:"http_addr"`
	GRPCAddr  string     `json:"grpc_addr"`
	Status    NodeStatus `json:"status"`
	LastSeen  time.Time  `json:"last_seen"`
	FailCount int        `json:"fail_count"`
}
