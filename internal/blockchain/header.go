package blockchain

/*
Header contains information about block
Params:
MerkleRoot - digest of merkle tree root
PreviousHash - digest of previous block's header
Holder - block author's hash address
TimeStamp - timestamp of block creation
*/
type Header struct {
	Version      string
	MerkleRoot   []byte
	PreviousHash []byte
	Holder       string
	TimeStamp    string

	// fields used in PoW consensus algorithm
	Difficulty uint8
	Nonce      uint8

	Signature []byte
}
