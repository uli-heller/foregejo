// Copyright 2024 The Forgejo Authors
// SPDX-License-Identifier: MIT

package actions

import (
	"path/filepath"
	"testing"

	"code.gitea.io/gitea/models/unittest"

	_ "code.gitea.io/gitea/models/actions"
	_ "code.gitea.io/gitea/models/activities"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m, &unittest.TestOptions{
		GiteaRootPath: filepath.Join("..", ".."),
	})
}
