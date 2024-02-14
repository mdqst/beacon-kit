// SPDX-License-Identifier: MIT
//
// Copyright (c) 2023 Berachain Foundation
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without
// restriction, including without limitation the rights to use,
// copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following
// conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sha256

import (
	"errors"

	"github.com/protolambda/ztyp/tree"
)

// SSZBytes represents a ssz-able uint64.
type SSZBytes []byte

// HashTreeRoot computes the hash tree root of the SSZBytes.
// It returns a fixed-size byte array and an error if any.
func (s SSZBytes) HashTreeRoot() ([32]byte, error) {
	return tree.GetHashFn().ByteListHTR(s, uint64(len(s))), nil
}

// We can visualize the process of building a Merkle tree as follows:
//
// [Element1] [Element2] ... [ElementN]
//
//	|          |                 |
//	v          v                 v
//
// [ Hash1 ]  [ Hash2 ]  ...  [ HashN ]  Hash each element
//
//	\         /                 /
//	 \       /       ...       /
//	  \     /                 /
//	   [       Merkle Tree       ]  Combine hashes to form the tree
//	             |
//	             v
//	         [ Root ]  The root hash of the Merkle tree
//
// BuildMerkleRoot constructs a Hash Tree Root (HTR) from a list of elements.
func BuildMerkleRoot[T Hashable](elements []T, limit uint64) (tree.Root, error) {
	roots, err := HashElements(elements)
	if err != nil {
		return [32]byte{}, err
	}
	return SafeMerkleizeVector(roots, limit)
}

// We can visualize the process of building a Merkle tree and mixing in the length as follows:
//
// [Element1] [Element2] ... [ElementN]
//
//	|          |                 |
//	v          v                 v
//
// [ Hash1 ]  [ Hash2 ]  ...  [ HashN ]  // Hash each element
//
//	\         /                 /
//	 \       /       ...       /
//	  \     /                 /
//	   [       Merkle Tree       ]  Combine hashes to form the tree
//	             |
//	             v
//	         [ Intermediate Root ]  The intermediate root hash of the Merkle tree
//	             |
//	             v
//	     [Intermediate Root + Length]  Append the length of the roots to the intermediate root
//	             |
//	             v
//	         [ Final Root ]  Hash the result to return the final HTR
//
// BuildMerkleRootAndMixinLength hashes each element in the list and then returns the HTR
// of the corresponding list of roots. It then appends the length of the roots to the
// end of the byteRoots and further hashes the result to return the final HTR.
func BuildMerkleRootAndMixinLength[T Hashable](elements []T, limit uint64) (tree.Root, error) {
	roots, err := HashElements(elements)
	if err != nil {
		return [32]byte{}, err
	}
	return SafeMerkelizeVectorAndMixinLength(roots, limit)
}

// BuildMerkleRootAndMixinLengthBytes hashes each element in the list and then returns the HTR.
func BuildMerkleRootAndMixinLengthBytes(elements [][]byte, limit uint64) (tree.Root, error) {
	roots, err := HashBytes(elements)
	if err != nil {
		return [32]byte{}, err
	}
	return SafeMerkelizeVectorAndMixinLength(roots, limit)
}

// HashElements hashes each element in the list and then returns each item as a
// tree.Root of height 1.
// The following diagram illustrates the process of hashing elements into tree roots:
//
// [Element1] -> Hash -> [Root1]
// [Element2] -> Hash -> [Root2]
//
//	.          .        .
//	.          .        .
//	.          .        .
//
// [ElementN] -> Hash -> [RootN]
//
// Where each Element is hashed individually to produce a corresponding Root.
// This process is applied to all elements in the input list, resulting in a list of roots.
func HashElements[T Hashable](elements []T) ([]tree.Root, error) {
	roots := make([]tree.Root, len(elements))
	var err error
	for i, el := range elements {
		roots[i], err = el.HashTreeRoot()
		if err != nil {
			return nil, err
		}
	}
	return roots, nil
}

// HashBytes hashes each element in the list and then returns each item as a
// tree.Root of height 1.
func HashBytes(elements [][]byte) ([]tree.Root, error) {
	roots := make([]tree.Root, len(elements))
	var err error
	for i, el := range elements {
		roots[i], err = SSZBytes(el).HashTreeRoot()
		if err != nil {
			return nil, err
		}
	}
	return roots, nil
}

// SafeMerkelizeVectorAndMixinLength takes a list of roots and returns the HTR
// of the corresponding list of roots. It then appends the length of the roots to the
// end of the byteRoots and further hashes the result to return the final HTR.
// The 'limit' parameter specifies the maximum allowed number of roots in the list,
// ensuring the list does not exceed this size.
func SafeMerkelizeVectorAndMixinLength(
	roots []tree.Root, maxRootsAllowed uint64,
) ([32]byte, error) {
	byteRoots, err := SafeMerkleizeVector(roots, maxRootsAllowed)
	if err != nil {
		return [32]byte{}, err
	}
	return tree.GetHashFn().Mixin(byteRoots, uint64(len(roots))), nil
}

// UnsafeMerkleizeVectorAndMixinLength is a function that operates on a list of tree roots.
// Initially, it computes the Hash Tree Root (HTR) for the given list. Subsequently, it
// appends the length of the list to the end of the computed byte array of roots. This
// appended byte array is then hashed again to produce the final HTR. This process can be
// visualized as follows:
//
// Step 1: Compute HTR for list of roots -> HTR([Root1, Root2, ..., RootN])
// Step 2: Append length of list to byte array -> [HTR_byte_array, length]
// Step 3: Hash the result from Step 2 -> HTR([HTR_byte_array, length])
//
// Given roots: [R1, R2, ..., RN]
// 1. Compute HTR -> [HTR_byte_array]
// 2. Append length -> [HTR_byte_array, length]
// Step 3: Hash result -> Final HTR.
func UnsafeMerkleizeVectorAndMixinLength(roots []tree.Root, maxRootsAllowed uint64) tree.Root {
	return tree.GetHashFn().Mixin(UnsafeMerkleizeVector(roots, uint64(len(roots))), maxRootsAllowed)
}

// UnsafeMerkleizeVector is a function that computes the Hash Tree Root (HTR) for
// a given list of tree roots. It simply calls the SafeMerkleizeVector function and
// panics if an error is returned.
func UnsafeMerkleizeVector(roots []tree.Root, maxRootsAllowed uint64) tree.Root {
	root, err := SafeMerkleizeVector(roots, maxRootsAllowed)
	if err != nil {
		panic(err)
	}
	return root
}

// The function SafeMerkleizeVector is designed to compute the Hash Tree Root (HTR)
// for a given list of tree roots. It operates under the assumption that no safety checks
// on the size of the list against a limit are needed (hence "Unsafe").
// Here's a step-by-step explanation and a diagrammatic representation of its operation:

// 1. Determine the depth required to cover the list, given a limit on the number of elements.
// 2. If the list is empty, return the zero hash at the calculated depth.
// 3. Iterate over each level of depth:
//    a. Check if the current list of roots has an odd length. If so, append a zero hash at
//       the current depth to make it even.
//    b. Hash pairs of elements (roots) together to form a new level of the tree, reducing
//       the total number of elements by half. This step is repeated until a single root is
//       obtained, representing the HTR of the list.

// Given roots: [R1, R2, R3]
// Step 3a: Check for odd length -> [R1, R2, R3, Z]
// Step 3b: Hash pairs -> [H(R1,R2), H(R3,Z)]
//
//	Repeat -> [H(H(R1,R2), H(R3,Z))]
//
// Result: The final HTR is H(H(R1,R2), H(R3,Z)).
func SafeMerkleizeVector(roots []tree.Root, maxRootsAllowed uint64) (tree.Root, error) {
	var err error

	// If the number of elements in the list exceeds the maximum allowed, return an error.
	if uint64(len(roots)) > maxRootsAllowed {
		return tree.Root{}, errors.New("merkleizing list exceeds the maximum allowed number of elements")
	}

	// Determine the max possible depth of the tree given maxRootsAllowed.
	depth := tree.CoverDepth(maxRootsAllowed)

	// Iterate over each level of depth in the tree.
	for i := uint8(0); i < depth; i++ {
		// If the current level of the tree has an odd number of roots, append the corresponding
		// zero hash for that depth to make it even.
		if len(roots)%2 != 0 {
			roots = append(roots, tree.ZeroHashes[i])
		} else if len(roots) == 0 {
			return tree.ZeroHashes[i], nil
		}

		// Hash pairs of elements together to form a new level of the tree.
		roots, err = HashTreeRoot(roots)
		if err != nil {
			return tree.Root{}, err
		}
	}
	return roots[0], nil
}
