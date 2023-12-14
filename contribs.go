package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/google/go-github/v57/github"
)

func main() {
	contribs, err := Get(context.Background(), os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}
	for _, c := range contribs {
		fmt.Printf("@%s\n", c)
	}
}

func Get(ctx context.Context, owner, repo string) ([]string, error) {
	contribs := make(map[string]bool)
	lp := github.ListOptions{PerPage: 1000}

	token := os.Getenv("GH")
	client := github.NewClient(nil).WithAuthToken(token)

	stars, _, err := client.Activity.ListStargazers(ctx, owner, repo, &lp)
	if err != nil {
		return nil, err
	}
	for _, s := range stars {
		contribs[*s.User.Login] = true
	}

	forks, _, err := client.Repositories.ListForks(ctx, owner, repo, &github.RepositoryListForksOptions{
		ListOptions: lp,
	})
	if err != nil {
		return nil, err
	}
	for _, f := range forks {
		contribs[*f.Owner.Login] = true
	}

	issues, _, err := client.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		ListOptions: lp,
	})
	if err != nil {
		return nil, err
	}

	for _, i := range issues {
		contribs[*i.User.Login] = true

		//	reacts, _, err := client.Reactions.ListIssueReactions(ctx, owner, repo, *i.Number, &lp)
		//	if err != nil {
		//		return nil, err
		//	}
		//	for _, r := range reacts {
		//		contribs[*r.User.Name] = true
		//	}

		// fetch comments
		comments, _, err := client.Issues.ListComments(ctx, owner, repo, *i.Number, &github.IssueListCommentsOptions{
			ListOptions: github.ListOptions{PerPage: 1000},
		})
		if err != nil {
			return nil, err
		}
		for _, c := range comments {
			contribs[*c.User.Login] = true

			//		reacts, _, err := client.Reactions.ListCommentReactions(ctx, owner, repo, *c.ID, &github.ListCommentReactionOptions{
			//			ListOptions: lp,
			//		})
			//		if err != nil {
			//			return nil, err
			//		}
			//		for _, r := range reacts {
			//			contribs[*r.User.Name] = true
			//		}
		}

		events, _, err := client.Issues.ListIssueEvents(ctx, owner, repo, *i.Number, &lp)
		if err != nil {
			return nil, err
		}
		for _, e := range events {
			contribs[*e.Actor.Login] = true
		}
	}
	prs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 1000},
	})
	if err != nil {
		return nil, err
	}
	for _, pr := range prs {
		contribs[*pr.User.Login] = true

		comments, _, err := client.PullRequests.ListComments(ctx, owner, repo, *pr.Number, &github.PullRequestListCommentsOptions{
			ListOptions: github.ListOptions{PerPage: 1000},
		})
		if err != nil {
			return nil, err
		}
		for _, c := range comments {
			contribs[*c.User.Login] = true
		}

		reviews, _, err := client.PullRequests.ListReviews(ctx, owner, repo, *pr.Number, &lp)
		if err != nil {
			return nil, err
		}
		for _, r := range reviews {
			contribs[*r.User.Login] = true
			comments, _, err := client.PullRequests.ListReviewComments(ctx, owner, repo, *pr.Number, *r.ID, &lp)
			if err != nil {
				return nil, err
			}
			for _, c := range comments {
				contribs[*c.User.Login] = true
			}
		}
	}
	var ret []string
	for c := range contribs {
		ret = append(ret, c)
	}
	sort.Strings(ret)
	return ret, nil
}
