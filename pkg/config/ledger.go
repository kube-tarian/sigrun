package config

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
)

type Ledger struct {
	Entries []*LedgerEntry
}

type GitCommitObject struct {
	Hash     string
	Message  string
	Author   string
	UnixTime int64
}

type LedgerEntry struct {
	Id          int
	Git         GitCommitObject
	Hash        string
	Timestamp   string
	Annotations map[string]string
	Checksum    *Checksum
}

type Checksum struct {
	Path     string
	Hash     string
	Children []*Checksum `json:"children,omitempty"`
}

func NewLedger() *Ledger {
	return &Ledger{}
}

func (l *Ledger) AddEntry(annotations map[string]string) error {
	checksum, err := NewChecksum(".")
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	headCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	l.Entries = append(l.Entries, &LedgerEntry{
		Id: len(l.Entries) + 1,
		Git: GitCommitObject{
			Hash:     headCommit.Hash.String(),
			Message:  headCommit.Message,
			Author:   headCommit.Author.Email,
			UnixTime: headCommit.Author.When.UnixNano(),
		},
		Hash:        checksum.Hash,
		Timestamp:   fmt.Sprint(time.Now().UnixNano()),
		Annotations: annotations,
		Checksum:    checksum,
	})

	return nil
}

func NewChecksum(path string) (*Checksum, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var children []*Checksum
	for _, f := range files {
		var checksum *Checksum
		currPath := filepath.Join(path, f.Name())

		if f.IsDir() {
			if filepath.Base(currPath) == ".git" {
				continue
			}

			checksum, err = NewChecksum(currPath)
			if err != nil {
				return nil, err
			}
		} else {
			fileR, err := os.Open(currPath)
			if err != nil {
				return nil, err
			}

			hasher := sha256.New()
			_, err = io.Copy(hasher, fileR)
			if err != nil {
				return nil, err
			}

			checksum = &Checksum{
				Path:     currPath,
				Hash:     base64.StdEncoding.EncodeToString(hasher.Sum(nil)),
				Children: nil,
			}
		}
		children = append(children, checksum)
	}

	hasher := sha256.New()
	var hashes []string
	for _, c := range children {
		hashes = append(hashes, c.Hash)
	}

	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i] > hashes[j]
	})

	_, err = io.Copy(hasher, strings.NewReader(strings.Join(hashes, "")))
	if err != nil {
		return nil, err
	}

	return &Checksum{
		Path:     path,
		Hash:     base64.StdEncoding.EncodeToString(hasher.Sum(nil)),
		Children: children,
	}, nil
}
