package node

type Service interface {
	ConnectPeer(address, url string) error
	DisconnectPeer(address string) error
	AcceptTransaction(data string) error
	AcceptBlock(data string) error
}
