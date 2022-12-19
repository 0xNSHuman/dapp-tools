package utils

import (
	"crypto/sha256"
	"errors"
	"math"

	"github.com/arriqaaq/merkletree"
)

func GenerateMerkleProof(
	treeData [][]byte, leaf []byte,
) ([][32]byte, error) {
	tree := merkletree.NewTree(treeData)

	var leafIndex uint64 = math.MaxUint64
	for i, dataItem := range treeData {
		if string(dataItem) == string(leaf) {
			leafIndex = uint64(i)
			break
		}
	}
	if leafIndex == math.MaxUint64 {
		return nil, errors.New("couldn't find the provided leaf in the tree data")
	}

	proof := tree.Proof(leafIndex)

	return [][sha256.Size]byte(proof), nil
}
