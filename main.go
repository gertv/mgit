package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"regexp"
)

type Repositories = chan RemoteRepo

func main() {
	log.Println("Reading config file")

	conffile := flag.String("config", "$HOME/.mgit/config.json", "mgit config file")

	config, err := ReadConfig(*conffile)
	if err != nil {
		log.Fatalf("Unable to read config file: %s", err)
	}

	repositories := scaffold(config)

	for repo := range repositories {
		for _, location := range config.Locations {
			if location.Wants(repo) {
				if err = Clone(repo, location); err != nil {
					log.Fatal(err)
				}
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

func ReadConfig(filename string) (config Config, err error) {
	file, err := os.Open(os.ExpandEnv(filename))
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

func (l Location) DirectoryName() string {
	vars := regexp.MustCompile("\\$[A-Z]*")

	return vars.ReplaceAllStringFunc(l.Directory, func(value string) string {
		log.Printf("Replacing " + value)
		return os.ExpandEnv(l.Directory)
	})
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