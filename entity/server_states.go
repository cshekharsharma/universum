package entity

const (
	SERVER_VERSION string = "v1.0.0-beta"

	STATE_STARTING      int32 = 1
	STATE_READY         int32 = 1 << 2
	STATE_SHUTTING_DOWN int32 = 1 << 8
	STATE_STOPPED       int32 = 1 << 16
)

var ServerState int32

func GetServerState() int32 {
	return ServerState
}

func GetServerStateAsString() string {
	switch ServerState {
	case STATE_STARTING:
		return "STARTING"

	case STATE_READY:
		return "READY"

	case STATE_SHUTTING_DOWN:
		return "SHUTTINGDOWN"

	case STATE_STOPPED:
		return "STOPPED"

	default:
		return "UNKNOWN"
	}
}
