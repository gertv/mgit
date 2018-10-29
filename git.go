package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

func Clone(repo RemoteRepo, location Location) error {
	fulldir := os.ExpandEnv(location.Directory)
	if stat, err := os.Stat(fulldir); err != nil {
		if os.IsNotExist(err) {
			log.Printf("Creating directory %s", fulldir)
			os.MkdirAll(fulldir, 0755)
		} else {
			return err
		}
	} else {
		if !stat.IsDir() {
			return errors.New(fmt.Sprintf("Unable to create directory %s - already exists", fulldir))
		}
	}

	if stat, err := os.Stat(path.Join(fulldir, repo.name)); err == nil {
		if stat.IsDir() {
			log.Printf("Repository location %s/%s already exists", fulldir, repo.name)
			// TODO: check if it the correct repo?
			return nil
		}
	}

	log.Printf("Cloning %s in %s/%s\n", repo.url, fulldir, repo.name)

	cmd := exec.Command("git", "clone", repo.url, repo.name)
	cmd.Dir = fulldir
	return cmd.Run()

}

func Fetch(repo LocalRepo) error {
	log.Printf("Fetching %s\n", repo.Directory)

	cmd := exec.Command("git", "fetch")
	cmd.Dir = repo.Directory
	return cmd.Run()
}