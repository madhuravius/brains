package repo_map

import "context"

func NewRepoMapConfig(ctx context.Context, repoRoot string) (RepoMapImpl, error) {
	r := &RepoMapConfig{}
	if err := r.BuildRepoMap(ctx, repoRoot); err != nil {
		return nil, err
	}
	return r, nil
}
