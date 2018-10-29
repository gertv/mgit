package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

type Repositories = chan RemoteRepo
type LocalRepositories = chan LocalRepo

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mgit",
		Short: "mgit is a small tool to work with multiple git repositories",
	}

	var conffile string
	rootCmd.PersistentFlags().StringVar(&conffile, "config", "$HOME/.mgit/config.json", "mgit config file")

	var cloneCmd = &cobra.Command{
		Use:   "clone",
		Short: "'git clone' all the repositories we defined",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ReadConfig(conffile)
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
		},
	}

	var fetchCmd = &cobra.Command{
		Use:   "fetch",
		Short: "'git fetch' on git repositories in all defined locations",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ReadConfig(conffile)
			if err != nil {
				log.Fatalf("Unable to read config file: %s", err)
			}

			repositories := localRepositories(config)

			for repo := range repositories {
				if err = Fetch(repo); err != nil {
					log.Printf("Error fetching %s: %s\n", repo.Directory, err)
				}
			}
		},
	}

	rootCmd.AddCommand(cloneCmd, fetchCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
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

func localRepositories(config Config) LocalRepositories {
	repositories := make(LocalRepositories)

	walkfunc := func(path string, info os.FileInfo, err error) error {
		if info.Name() == ".git" {
			repositories <- LocalRepo{filepath.Dir(path)}
			return filepath.SkipDir
		}
		return nil
	}

	go func() {
		defer close(repositories)

		for _, location := range config.Locations {
			err := filepath.Walk(os.ExpandEnv(location.Directory), walkfunc)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return repositories
}

func ReadConfig(filename string) (config Config, err error) {
	log.Printf("Reading config file %s", filename)
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

type LocalRepo struct {
	Directory string
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