package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/libgit2/git2go"
)

type repoTestFunc func(conf *git.Config)

func withinStubRepo(t *testing.T, repoPath string, repoTest repoTestFunc) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	conf, err := initializeRepo(repoPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repoPath)

	err = os.Chdir(repoPath)
	if err != nil {
		t.Fatal(err)
	}

	repoTest(conf)

	err = os.Chdir(cwd)
	if err != nil {
		t.Fatal(err)
	}
}

func mockHomeEnv(dir string) {
	cwd, err := os.Getwd()
	if err != nil {
		os.Stderr.WriteString("Could not get current directory\n")
		os.Exit(2)
	}

	dir = path.Join(cwd, dir)
	_, err = os.Open(dir)
	if err != nil {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			os.Stderr.WriteString("Could not create directory\n")
			os.Exit(2)
		}
	}

	os.Setenv("HOME", dir)
}

func initializeRepo(p string) (*git.Config, error) {
	repo, err := git.InitRepository(p, true)
	if err != nil {
		return nil, err
	}

	conf, err := repo.Config()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func closeFile(f *os.File) {
	name := f.Name()
	f.Close()
	os.Remove(name)
}

func initTestGitConfig(path string, t *testing.T) *git.Config {
	gitconfig, err := git.NewConfig()
	if err != nil {
		t.Error(err)
	}

	err = gitconfig.AddFile(path, git.ConfigLevelHighest, false)
	if err != nil {
		t.Error(err)
	}

	return gitconfig
}

func createPearrc(t *testing.T, contents []byte) *os.File {
	p := path.Join(os.Getenv("HOME"), ".pearrc")
	f, err := os.Create(p)
	if err != nil {
		os.Stdout = os.Stderr
		t.Fatalf("Could not create .pearrc %s", err)
	}

	_, err = f.Write(contents)
	if err != nil {
		os.Stdout = os.Stderr
		t.Fatal("Could not write to .pearrc %s", err)
	}

	return f
}

func mockStdin(t *testing.T, contents string) *os.File {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmp.WriteString(contents + "\n")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmp.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdin = tmp

	return tmp
}

func mockStdout(t *testing.T) (*os.File, *os.File) {
	oldstdout := os.Stdout
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error("Could not create temp file")
	}

	os.Stdout = tmp

	return tmp, oldstdout
}

func cleanupStdout(t *testing.T, tmp *os.File, stdout *os.File) {
	err := tmp.Close()
	if err != nil {
		t.Error(err)
	}

	os.Stdout = stdout
}

func restorePearrc(t *testing.T, contents []byte) {
	p := path.Join(os.Getenv("HOME"), ".pearrc")
	err := ioutil.WriteFile(p, contents, os.ModeExclusive)
	if err != nil {
		t.Error(err)
	}
}
