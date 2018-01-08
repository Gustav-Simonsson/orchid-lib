package p2p

const (
	// Default in io.Copy is 30 * 1024
	// On source nodes this is probably fine, on relays and exits
	// (especially if embedded devices) this can be important to tune
	// Used by TCPProxy and DCReadWriteCloser
	transferBufSize = 30 * 1024
)
