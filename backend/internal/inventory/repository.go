package inventory

import "context"

// ServerRepository persists and loads Server entities.
type ServerRepository interface {
	GetByID(ctx context.Context, id string) (*Server, error)
	List(ctx context.Context) ([]Server, error)
	Create(ctx context.Context, server *Server) error
	Update(ctx context.Context, server *Server) error
	Delete(ctx context.Context, id string) error
}

// ClusterRepository persists and loads Cluster entities.
type ClusterRepository interface {
	GetByID(ctx context.Context, id string) (*Cluster, error)
	List(ctx context.Context) ([]Cluster, error)
	Create(ctx context.Context, cluster *Cluster) error
	Update(ctx context.Context, cluster *Cluster) error
	Delete(ctx context.Context, id string) error
}

// DatabaseRepository persists and loads Database entities.
type DatabaseRepository interface {
	GetByID(ctx context.Context, id string) (*Database, error)
	ListByClusterID(ctx context.Context, clusterID string) ([]Database, error)
	Create(ctx context.Context, database *Database) error
	Update(ctx context.Context, database *Database) error
	Delete(ctx context.Context, id string) error
}

// AgentRepository persists and loads Agent entities.
type AgentRepository interface {
	GetByID(ctx context.Context, id string) (*Agent, error)
	GetByServerID(ctx context.Context, serverID string) (*Agent, error)
	Create(ctx context.Context, agent *Agent) error
	Update(ctx context.Context, agent *Agent) error
	Delete(ctx context.Context, id string) error
}
