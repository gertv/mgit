package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
)

type Repositories = chan RemoteRepo

func main() {
	log.Println("Reading config file")

	config, err := ReadConfig()
	if err != nil {
		log.Fatalf("Unable to read config file: %s", err)
	}

	repositories := scaffold(config)

	for repo := range repositories {
		for _, location := range config.Locations {
			if location.Wants(repo) {
				log.Printf("We should clone %s in %s/%s\n", repo.url, location.Directory, repo.name)
			}
		}
	}
}

func scaffold(config Config) Repositories {
	repositories := make(Repositories)

	github := createGithubSource(config)
	go func() {
		defer close(repositories)

		github.scaffold(repositories)
	}()

	deduped := make(Repositories)
	go func() {
		defer close(deduped)

		seen := make(map[string]bool)
		for repo := range repositories {
			value, found := seen[repo.url]
			if !found || !value {
				deduped <- repo
				seen[repo.url] = true
			}
		}
	}()

	return deduped
}

func ReadConfig() (config Config, err error) {
	file, err := os.Open("mgit.json")
	if err != nil {
		return
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return
	}

	return
}

type Location struct {
	Directory string `json:"directory"`
	Repository string `json:"repository"`
}

func (l Location) Wants(repo RemoteRepo) bool {
	match, err := regexp.MatchString(l.Repository, repo.url)
	if err != nil {
		log.Printf("Unable to match regex '%s': %s\n", l.Repository, err)
		return false
	}
	return match
}

type Github struct {
	Token string `json:"token"`
}

type Sources struct {
	Github Github `json:"github"`
}

type Config struct {
	Sources Sources `json:"sources"`
	Locations []Location  `json:"locations"`
}

type RemoteRepo struct {
	name string
	url string
}