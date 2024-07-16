package types

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/cbergoon/merkletree"
	pb "google.golang.org/protobuf/proto"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{
		hash: hash,
	}
}

func (t TxHash) CalculateHash() ([]byte, error) {
	return t.hash, nil
}

func (t TxHash) Equals(other merkletree.Content) (bool, error) {

	return bytes.Equal(t.hash, other.(TxHash).hash), nil
}

func VerifyBlock(b *proto.Block) error {
	if len(b.Transactions) > 0 {
		if err := VerifyRootHash(b); err != nil {
			return err
		}
	}

	if len(b.PublicKey) != crypto.PubKeyLen {
		return errors.New("invalid public key")
	}

	if len(b.Signature) != crypto.SignatureLen {
		return errors.New("invalid signature")
	}

	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	hash := HashBlock(b)
	if !sig.Verify(pubKey, hash) {
		return errors.New("invalid block: can't verify signature")
	}
	return nil
}

func SignBlock(pk crypto.PrivateKey, b *proto.Block) crypto.Signature {
	hash := HashBlock(b)
	sign := pk.Sign(hash)

	b.PublicKey = pk.Public().Bytes()
	b.Signature = sign.Bytes()

	if len(b.Transactions) > 0 {
		tree, err := GetMerkleTree(b)
		if err != nil {
			panic(err)
		}
		b.Header.RootHash = tree.MerkleRoot()
	}

	return sign
}

func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func VerifyRootHash(b *proto.Block) error {
	t, err := GetMerkleTree(b)
	if err != nil {
		return err
	}

	ok, err := t.VerifyTree()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("invalid root hash")
	}

	if !bytes.Equal(b.Header.RootHash, t.MerkleRoot()) {
		return errors.New("invalid merkle root hash")
	}

	return nil
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, 0, len(b.Transactions))
	for _, tx := range b.Transactions {
		list = append(list, NewTxHash(HashTransaction(tx)))
	}
	t, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}
