package command

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/remote"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/terraform"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/mitchellh/cli"
)

func TestPull_noRemote(t *testing.T) {
	tmp, cwd := testCwd(t)
	defer testFixCwd(t, tmp, cwd)

	ui := new(cli.MockUi)
	c := &PullCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}

	args := []string{}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: \n%s", ui.ErrorWriter.String())
	}
}

func TestPull_local(t *testing.T) {
	tmp, cwd := testCwd(t)
	defer testFixCwd(t, tmp, cwd)

	s := terraform.NewState()
	s.Serial = 10
	conf, srv := testRemoteState(t, s, 200)

	s = terraform.NewState()
	s.Serial = 5
	s.Remote = conf
	defer srv.Close()

	// Store the local state
	buf := bytes.NewBuffer(nil)
	terraform.WriteState(s, buf)
	remote.EnsureDirectory()
	remote.Persist(buf)

	ui := new(cli.MockUi)
	c := &PullCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}
	args := []string{}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: \n%s", ui.ErrorWriter.String())
	}
}

// testRemoteState is used to make a test HTTP server to
// return a given state file
func testRemoteState(t *testing.T, s *terraform.State, c int) (*terraform.RemoteState, *httptest.Server) {
	var b64md5 string
	buf := bytes.NewBuffer(nil)

	if s != nil {
		enc := json.NewEncoder(buf)
		if err := enc.Encode(s); err != nil {
			t.Fatalf("err: %v", err)
		}
		md5 := md5.Sum(buf.Bytes())
		b64md5 = base64.StdEncoding.EncodeToString(md5[:16])
	}

	cb := func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "PUT" {
			resp.WriteHeader(c)
			return
		}
		if s == nil {
			resp.WriteHeader(404)
			return
		}
		resp.Header().Set("Content-MD5", b64md5)
		resp.Write(buf.Bytes())
	}
	srv := httptest.NewServer(http.HandlerFunc(cb))
	remote := &terraform.RemoteState{
		Type:   "http",
		Config: map[string]string{"address": srv.URL},
	}
	return remote, srv
}
