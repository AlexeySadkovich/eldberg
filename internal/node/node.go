package node

type NodeService interface {
	ConnectPeer(address, url string) error
	DisconnectPeer(address string)
	AcceptTransaction(data string) error
	AcceptBlock(data string) error
}
