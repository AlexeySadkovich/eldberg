package node

type Service interface {
	ConnectPeer(address, url string) error
	DisconnectPeer(address string) error
	AcceptTransaction(data []byte) error
	AcceptBlock(data []byte) error
}
