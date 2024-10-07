// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2024, Berachain Foundation. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

//nolint:lll // long strings
package hex_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/berachain/beacon-kit/mod/primitives/pkg/encoding/hex"
	"github.com/stretchr/testify/require"
)

// ====================== Constructors ===========================.
func TestNewStringInvariants(t *testing.T) {
	// NewString constructor should never error or panic
	// output should always satisfy the string invariants regardless of input
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Valid hex string",
			input: "0x48656c6c6f",
		},
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "No 0x prefix",
			input: "48656c6c6f",
		},
		{
			name:  "Valid single hex character",
			input: "0x0",
		},
		{
			name:  "Empty hex string",
			input: "0x",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			str := hex.NewString(test.input)
			_, err := hex.IsValidHex(str)
			require.NoError(t, err)
		})
	}
}

// ====================== Numeric ===========================.

// FromBigInt, then ToBigInt.
func TestBigIntRoundTrip(t *testing.T) {
	// assume FromBigInt only called on non-negative big.Int
	tests := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{
			name:     "zero value",
			input:    big.NewInt(0),
			expected: "0x0",
		},
		{
			name:     "positive value",
			input:    big.NewInt(12345),
			expected: "0x3039",
		},
		{
			name:     "large positive value",
			input:    new(big.Int).SetBytes(bytes.Repeat([]byte{0xff}, 32)),
			expected: "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hex.FromBigInt(tt.input)
			require.Equal(t, tt.expected, result.Unwrap())

			_, err := hex.IsValidHex(result)
			require.NoError(t, err)

			var dec *big.Int

			if tt.input.Sign() >= 0 {
				dec, err = hex.NewString(result.Unwrap()).ToBigInt()
			} else {
				dec, err = hex.NewString(result.Unwrap()).ToBigInt()
				dec = dec.Neg(dec)
			}

			require.NoError(t, err)
			require.Zero(t, dec.Cmp(tt.input))
		})
	}
}

// ====================== Helpers ===========================.

func TestString_MustToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    hex.String
		expected *big.Int
		panics   bool
	}{
		{"Valid hex string", "0x1", big.NewInt(1), false},
		{"Another valid hex string", "0x10", big.NewInt(16), false},
		{"Large valid hex string", "0x1a", big.NewInt(26), false},
		{"Invalid hex string", "0xinvalid", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				res *big.Int
				f   = func() {
					res = tt.input.MustToBigInt()
				}
			)
			if tt.panics {
				require.Panics(t, f)
			} else {
				require.NotPanics(t, f)
				require.Equal(t, tt.expected, res)
			}
		})
	}
}
