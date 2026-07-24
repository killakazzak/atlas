// Package inventory defines the inventory domain model and persistence ports.
// It contains entities and repository interfaces only — no storage or HTTP.
package inventory

import "time"

// ServerStatus is the operational state of a managed host.
type ServerStatus string

// Known ServerStatus values.
const (
	ServerStatusUnknown     ServerStatus = "unknown"
	ServerStatusOnline      ServerStatus = "online"
	ServerStatusOffline     ServerStatus = "offline"
	ServerStatusDegraded    ServerStatus = "degraded"
	ServerStatusMaintenance ServerStatus = "maintenance"
)

// AgentStatus is the connectivity state of an Atlas agent.
type AgentStatus string

// Known AgentStatus values.
const (
	AgentStatusUnknown AgentStatus = "unknown"
	AgentStatusActive  AgentStatus = "active"
	AgentStatusStale   AgentStatus = "stale"
	AgentStatusOffline AgentStatus = "offline"
)

// Server is a managed host in the inventory.
type Server struct {
	ID              string
	Name            string
	Hostname        string
	IP              string
	OperatingSystem string
	Status          ServerStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Cluster groups related servers running a 1C platform version.
type Cluster struct {
	ID              string
	Name            string
	PlatformVersion string
	Servers         []Server
}

// Database is an infobase or DBMS instance belonging to a cluster.
type Database struct {
	ID        string
	Name      string
	DBMS      string
	ClusterID string
}

// Agent is the Atlas agent installed on a server.
type Agent struct {
	ID       string
	ServerID string
	Version  string
	LastSeen time.Time
	Status   AgentStatus
}
