package node

import "time"

// Config set for ssr config
type Config struct {
	Server        string `json:"server"`
	ServerPort    int    `json:"server_port"`
	Method        string `json:"method"`
	Protocol      string `json:"protocol"`
	ProtocolParam string `json:"protocol_param"`
	OBFS          string `json:"obfs"`
	OBFSParam     string `json:"obfs_param"`
	Password      string `json:"password"`
	Remarks       string `json:"remarks"`
	Group         string `json:"group"`

	ID        string    `json:"id"`
	CheckTime time.Time `json:"check_time"`
}

type CheckConfig struct {
	TCPTimeout string `yaml:"timeout"`
	Not        string `yaml:"not"`
}
