// Package pgtest provides a convenient way of spinning up a postgres server
// for tests. Avoid if at all possible -- best left for use by slow integration
// tests. Start/Stop panic on errors, making this ideally suited for use via
// TestMain where one doesn't have a *testing.T available.
package pgtest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

var conf = `
fsync = off
listen_addresses = ''
max_connections = 100
unix_socket_directories = '%s'
shared_buffers = 128MB
timezone = 'UTC'
`

type Server struct {
	URL string
	dir string
	cmd *exec.Cmd
}

// Start a new postgres instance.
func Start() *Server {
	out, err := exec.Command("pg_config", "--bindir").Output()
	if err != nil {
		panic(err)
	}
	bindir := string(bytes.TrimSpace(out))
	postgres := filepath.Join(bindir, "postgres")
	initdb := filepath.Join(bindir, "initdb")
	dir, err := ioutil.TempDir(
		"", fmt.Sprintf("pgtest-%s", filepath.Base(os.Args[0])))
	if err != nil {
		panic(err)
	}
	s := &Server{dir: dir}
	if err := exec.Command(initdb, "-D", dir).Run(); err != nil {
		os.RemoveAll(dir)
		panic(err)
	}
	err = ioutil.WriteFile(
		filepath.Join(dir, "postgresql.conf"),
		[]byte(fmt.Sprintf(conf, dir)),
		0666,
	)
	if err != nil {
		os.RemoveAll(dir)
		panic(err)
	}
	s.URL = "host=" + dir + " dbname=postgres sslmode=disable"
	s.cmd = exec.Command(postgres, "-D", dir)
	if err := s.cmd.Start(); err != nil {
		os.RemoveAll(dir)
		panic(err)
	}
	sock := filepath.Join(dir, ".s.PGSQL.5432")
	for n := 0; n < 20; n++ {
		if _, err := os.Stat(sock); err == nil {
			return s
		}
		time.Sleep(50 * time.Millisecond)
	}
	os.RemoveAll(dir)
	panic("timeout waiting for postgres to start")
}

// Stop the postgres server and remove the data directory.
func (s *Server) Stop() {
	if err := s.cmd.Process.Signal(syscall.SIGKILL); err != nil {
		os.RemoveAll(s.dir)
		panic(err)
	}
	if err := os.RemoveAll(s.dir); err != nil {
		panic(err)
	}
}
