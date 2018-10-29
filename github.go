package main

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
)

type GithubSource struct {
	context context.Context
	client *github.Client
}

func createGithubSource(config Config) GithubSource {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Sources.Github.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return GithubSource{ctx, github.NewClient(tc)}
}

func (source GithubSource) scaffold(repos Repositories) {
	log.Println("Scaffolding Github repositories")
	source.readUserRepositories(repos)
	source.readWatchedRepositories(repos)
	source.readStarredRepositories(repos)
}

func (source GithubSource) readUserRepositories(repos Repositories) {
	context := source.context
	client := source.client

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 50},
	}

	for {
		list, resp, err := client.Repositories.List(context, "", opt)
		if err != nil {
			log.Printf("Error reading for ...: %s", err)
		}
		for _, repo := range list {
			repos <- RemoteRepo{*repo.Name, *repo.SSHURL}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}

func (source GithubSource) readWatchedRepositories(repos Repositories) {
	context := source.context
	client := source.client

	opt := &github.ListOptions{PerPage: 50}

	for {
		list, resp, err := client.Activity.ListWatched(context, "", opt)
		if err != nil {
			log.Printf("Error reading for ...: %s", err)
		}
		for _, repo := range list {
			repos <- RemoteRepo{*repo.Name, *repo.SSHURL}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}

func (source GithubSource) readStarredRepositories(repos Repositories) {
	context := source.context
	client := source.client

	opt := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: 50},
	}

	for {
		list, resp, err := client.Activity.ListStarred(context, "", opt)
		if err != nil {
			log.Printf("Error reading for ...: %s", err)
		}
		for _, repo := range list {
			repos <- RemoteRepo{*repo.Repository.Name, *repo.Repository.SSHURL}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}