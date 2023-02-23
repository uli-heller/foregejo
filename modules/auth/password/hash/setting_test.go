// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckSettingPasswordHashAlgorithm(t *testing.T) {
	t.Run("pbkdf2 is pbkdf2_v2", func(t *testing.T) {
		pbkdf2v2Config, pbkdf2v2Algo := SetDefaultPasswordHashAlgorithm("pbkdf2_v2")
		pbkdf2Config, pbkdf2Algo := SetDefaultPasswordHashAlgorithm("pbkdf2")

		assert.Equal(t, pbkdf2v2Config, pbkdf2Config)
		assert.Equal(t, pbkdf2v2Algo.Name, pbkdf2Algo.Name)
	})

	for a, b := range aliasAlgorithmNames {
		t.Run(a+"="+b, func(t *testing.T) {
			aConfig, aAlgo := SetDefaultPasswordHashAlgorithm(a)
			bConfig, bAlgo := SetDefaultPasswordHashAlgorithm(b)

			assert.Equal(t, bConfig, aConfig)
			assert.Equal(t, aAlgo.Name, bAlgo.Name)
		})
	}

	t.Run("pbkdf2_hi is the default when default password hash algorithm is empty", func(t *testing.T) {
		emptyConfig, emptyAlgo := SetDefaultPasswordHashAlgorithm("")
		pbkdf2hiConfig, pbkdf2hiAlgo := SetDefaultPasswordHashAlgorithm("pbkdf2_hi")

		assert.Equal(t, pbkdf2hiConfig, emptyConfig)
		assert.Equal(t, pbkdf2hiAlgo.Name, emptyAlgo.Name)
	})
}
