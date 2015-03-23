package types


type SharedMaptype  map[int]chan int

//state
const (
	FOLLOWER = "follower"
	CANDIDATE = "candidate"
	LEADER = "leader"
)

//event types
const (
	ClientAppend = 1
	VoteRequest = 2
	AppendRPC = 3
	Timeout = 4
)

const (
	UnknownLeader = 0
)

// --------------------------------------
// Raft setup
type ServerConfig struct {
	Id         int    // Id of server. Must be unique
	Hostname   string // name or ip of host
	ClientPort int    // port at which server listens to client messages.
	LogPort    int    // tcp port for inter-replica protocol messages.
}

type ClusterConfig struct {
	Path    string         // Directory for persistent log
	Servers []ServerConfig // All servers in this cluster
}
