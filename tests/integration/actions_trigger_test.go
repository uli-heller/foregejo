// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/url"
	"strings"
	"testing"
	"time"

	actions_model "code.gitea.io/gitea/models/actions"
	"code.gitea.io/gitea/models/db"
	git_model "code.gitea.io/gitea/models/git"
	issues_model "code.gitea.io/gitea/models/issues"
	repo_model "code.gitea.io/gitea/models/repo"
	unit_model "code.gitea.io/gitea/models/unit"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	actions_module "code.gitea.io/gitea/modules/actions"
	"code.gitea.io/gitea/modules/git"
	webhook_module "code.gitea.io/gitea/modules/webhook"
	actions_service "code.gitea.io/gitea/services/actions"
	pull_service "code.gitea.io/gitea/services/pull"
	release_service "code.gitea.io/gitea/services/release"
	repo_service "code.gitea.io/gitea/services/repository"
	files_service "code.gitea.io/gitea/services/repository/files"

	"github.com/stretchr/testify/assert"
)

func TestPullRequestTargetEvent(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, u *url.URL) {
		user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2}) // owner of the base repo
		org3 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 3})  // owner of the forked repo

		// create the base repo
		baseRepo, err := repo_service.CreateRepository(db.DefaultContext, user2, user2, repo_service.CreateRepoOptions{
			Name:          "repo-pull-request-target",
			Description:   "test pull-request-target event",
			AutoInit:      true,
			Gitignores:    "Go",
			License:       "MIT",
			Readme:        "Default",
			DefaultBranch: "main",
			IsPrivate:     false,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, baseRepo)

		// enable actions
		err = repo_model.UpdateRepositoryUnits(baseRepo, []repo_model.RepoUnit{{
			RepoID: baseRepo.ID,
			Type:   unit_model.TypeActions,
		}}, nil)
		assert.NoError(t, err)

		// create the forked repo
		forkedRepo, err := repo_service.ForkRepository(git.DefaultContext, user2, org3, repo_service.ForkRepoOptions{
			BaseRepo:    baseRepo,
			Name:        "forked-repo-pull-request-target",
			Description: "test pull-request-target event",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, forkedRepo)

		// add workflow file to the base repo
		addWorkflowToBaseResp, err := files_service.ChangeRepoFiles(git.DefaultContext, baseRepo, user2, &files_service.ChangeRepoFilesOptions{
			Files: []*files_service.ChangeRepoFile{
				{
					Operation:     "create",
					TreePath:      ".gitea/workflows/pr.yml",
					ContentReader: strings.NewReader("name: test\non:\n  pull_request_target:\n    paths:\n      - 'file_*.txt'\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo helloworld\n"),
				},
			},
			Message:   "add workflow",
			OldBranch: "main",
			NewBranch: "main",
			Author: &files_service.IdentityOptions{
				Name:  user2.Name,
				Email: user2.Email,
			},
			Committer: &files_service.IdentityOptions{
				Name:  user2.Name,
				Email: user2.Email,
			},
			Dates: &files_service.CommitDateOptions{
				Author:    time.Now(),
				Committer: time.Now(),
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, addWorkflowToBaseResp)

		// add a new file to the forked repo
		addFileToForkedResp, err := files_service.ChangeRepoFiles(git.DefaultContext, forkedRepo, org3, &files_service.ChangeRepoFilesOptions{
			Files: []*files_service.ChangeRepoFile{
				{
					Operation:     "create",
					TreePath:      "file_1.txt",
					ContentReader: strings.NewReader("file1"),
				},
			},
			Message:   "add file1",
			OldBranch: "main",
			NewBranch: "fork-branch-1",
			Author: &files_service.IdentityOptions{
				Name:  org3.Name,
				Email: org3.Email,
			},
			Committer: &files_service.IdentityOptions{
				Name:  org3.Name,
				Email: org3.Email,
			},
			Dates: &files_service.CommitDateOptions{
				Author:    time.Now(),
				Committer: time.Now(),
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, addFileToForkedResp)

		// create Pull
		pullIssue := &issues_model.Issue{
			RepoID:   baseRepo.ID,
			Title:    "Test pull-request-target-event",
			PosterID: org3.ID,
			Poster:   org3,
			IsPull:   true,
		}
		pullRequest := &issues_model.PullRequest{
			HeadRepoID: forkedRepo.ID,
			BaseRepoID: baseRepo.ID,
			HeadBranch: "fork-branch-1",
			BaseBranch: "main",
			HeadRepo:   forkedRepo,
			BaseRepo:   baseRepo,
			Type:       issues_model.PullRequestGitea,
		}
		err = pull_service.NewPullRequest(git.DefaultContext, baseRepo, pullIssue, nil, nil, pullRequest, nil)
		assert.NoError(t, err)
		// if a PR "synchronized" event races the "opened" event by having the same SHA, it must be skipped. See https://codeberg.org/forgejo/forgejo/issues/2009.
		assert.True(t, actions_service.SkipPullRequestEvent(git.DefaultContext, webhook_module.HookEventPullRequestSync, baseRepo.ID, addFileToForkedResp.Commit.SHA))

		// load and compare ActionRun
		assert.Equal(t, 1, unittest.GetCount(t, &actions_model.ActionRun{RepoID: baseRepo.ID}))
		actionRun := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{RepoID: baseRepo.ID})
		assert.Equal(t, addFileToForkedResp.Commit.SHA, actionRun.CommitSHA)
		assert.Equal(t, actions_module.GithubEventPullRequestTarget, actionRun.TriggerEvent)

		// add another file whose name cannot match the specified path
		addFileToForkedResp, err = files_service.ChangeRepoFiles(git.DefaultContext, forkedRepo, org3, &files_service.ChangeRepoFilesOptions{
			Files: []*files_service.ChangeRepoFile{
				{
					Operation:     "create",
					TreePath:      "foo.txt",
					ContentReader: strings.NewReader("foo"),
				},
			},
			Message:   "add foo.txt",
			OldBranch: "main",
			NewBranch: "fork-branch-2",
			Author: &files_service.IdentityOptions{
				Name:  org3.Name,
				Email: org3.Email,
			},
			Committer: &files_service.IdentityOptions{
				Name:  org3.Name,
				Email: org3.Email,
			},
			Dates: &files_service.CommitDateOptions{
				Author:    time.Now(),
				Committer: time.Now(),
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, addFileToForkedResp)

		// create Pull
		pullIssue = &issues_model.Issue{
			RepoID:   baseRepo.ID,
			Title:    "A mismatched path cannot trigger pull-request-target-event",
			PosterID: org3.ID,
			Poster:   org3,
			IsPull:   true,
		}
		pullRequest = &issues_model.PullRequest{
			HeadRepoID: forkedRepo.ID,
			BaseRepoID: baseRepo.ID,
			HeadBranch: "fork-branch-2",
			BaseBranch: "main",
			HeadRepo:   forkedRepo,
			BaseRepo:   baseRepo,
			Type:       issues_model.PullRequestGitea,
		}
		err = pull_service.NewPullRequest(git.DefaultContext, baseRepo, pullIssue, nil, nil, pullRequest, nil)
		assert.NoError(t, err)

		// the new pull request cannot trigger actions, so there is still only 1 record
		assert.Equal(t, 1, unittest.GetCount(t, &actions_model.ActionRun{RepoID: baseRepo.ID}))
	})
}

func TestCreateDeleteRefEvent(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, u *url.URL) {
		user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

		// create the repo
		repo, err := repo_service.CreateRepository(db.DefaultContext, user2, user2, repo_service.CreateRepoOptions{
			Name:          "create-delete-ref-event",
			Description:   "test create delete ref ci event",
			AutoInit:      true,
			Gitignores:    "Go",
			License:       "MIT",
			Readme:        "Default",
			DefaultBranch: "main",
			IsPrivate:     false,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, repo)

		// enable actions
		err = repo_model.UpdateRepositoryUnits(repo, []repo_model.RepoUnit{{
			RepoID: repo.ID,
			Type:   unit_model.TypeActions,
		}}, nil)
		assert.NoError(t, err)

		// reload units
		repo.Units = nil
		assert.NoError(t, repo.LoadUnits(db.DefaultContext))

		// add workflow file to the repo
		addWorkflowToBaseResp, err := files_service.ChangeRepoFiles(git.DefaultContext, repo, user2, &files_service.ChangeRepoFilesOptions{
			Files: []*files_service.ChangeRepoFile{
				{
					Operation:     "create",
					TreePath:      ".gitea/workflows/createdelete.yml",
					ContentReader: strings.NewReader("name: test\non:\n  [create,delete]\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo helloworld\n"),
				},
			},
			Message:   "add workflow",
			OldBranch: "main",
			NewBranch: "main",
			Author: &files_service.IdentityOptions{
				Name:  user2.Name,
				Email: user2.Email,
			},
			Committer: &files_service.IdentityOptions{
				Name:  user2.Name,
				Email: user2.Email,
			},
			Dates: &files_service.CommitDateOptions{
				Author:    time.Now(),
				Committer: time.Now(),
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, addWorkflowToBaseResp)

		// Get the commit ID of the default branch
		gitRepo, err := git.OpenRepository(git.DefaultContext, repo_model.RepoPath(user2.Name, repo.Name))
		assert.NoError(t, err)
		defer gitRepo.Close()
		branch, err := git_model.GetBranch(db.DefaultContext, repo.ID, repo.DefaultBranch)
		assert.NoError(t, err)

		// create a branch
		err = repo_service.CreateNewBranchFromCommit(db.DefaultContext, user2, repo, gitRepo, branch.CommitID, "test-create-branch")
		assert.NoError(t, err)
		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{
			Title:      "add workflow",
			RepoID:     repo.ID,
			Event:      "create",
			Ref:        "refs/heads/test-create-branch",
			WorkflowID: "createdelete.yml",
			CommitSHA:  branch.CommitID,
		})
		assert.NotNil(t, run)

		// create a tag
		err = release_service.CreateNewTag(db.DefaultContext, user2, repo, branch.CommitID, "test-create-tag", "test create tag event")
		assert.NoError(t, err)
		run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{
			Title:      "add workflow",
			RepoID:     repo.ID,
			Event:      "create",
			Ref:        "refs/tags/test-create-tag",
			WorkflowID: "createdelete.yml",
			CommitSHA:  branch.CommitID,
		})
		assert.NotNil(t, run)

		// delete the branch
		err = repo_service.DeleteBranch(db.DefaultContext, user2, repo, gitRepo, "test-create-branch")
		assert.NoError(t, err)
		run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{
			Title:      "add workflow",
			RepoID:     repo.ID,
			Event:      "delete",
			Ref:        "main",
			WorkflowID: "createdelete.yml",
			CommitSHA:  branch.CommitID,
		})
		assert.NotNil(t, run)

		// delete the tag
		tag, err := repo_model.GetRelease(db.DefaultContext, repo.ID, "test-create-tag")
		assert.NoError(t, err)
		err = release_service.DeleteReleaseByID(db.DefaultContext, repo, tag, user2, true)
		assert.NoError(t, err)
		run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{
			Title:      "add workflow",
			RepoID:     repo.ID,
			Event:      "delete",
			Ref:        "main",
			WorkflowID: "createdelete.yml",
			CommitSHA:  branch.CommitID,
		})
		assert.NotNil(t, run)
	})
}
