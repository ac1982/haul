package bilibili

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/ac1982/haul/internal/testkit"
)

// TestMain points the config directory at a temporary one, so no test reads or writes the real logins. Tests that
// write logins there run sequentially and remove what they wrote before the parallel ones start.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "haul-bilibili-")
	if err != nil {
		panic(err)
	}
	os.Setenv("HAUL_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

var ctx = context.Background()

// fixture answers with a captured response from testdata.
func fixture(t *testing.T, name string) testkit.Handler {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return func(*http.Request) testkit.Response { return testkit.JSON(string(b)) }
}

func answer(text string) testkit.Handler {
	return func(*http.Request) testkit.Response { return testkit.JSON(text) }
}

func status(code int) testkit.Handler {
	return func(*http.Request) testkit.Response { return testkit.Status(code) }
}

// stubbed is an extractor over a fresh stub network with default options changed by configure.
func stubbed(configure func(*Options)) (*Extractor, *testkit.Stub) {
	stub := testkit.NewStub()
	o := DefaultOptions()
	if configure != nil {
		configure(&o)
	}
	return New(stub.Client(), o), stub
}

// plainSession is a logged-out session with the default hosts; tests that skip Resolve use it.
func plainSession() *session {
	return &session{userAgent: "UA", host: defaultHost, epHost: defaultHost, tvHost: defaultTVHost}
}

// stubCaptured registers the logged-out responses captured from the live API.
func stubCaptured(t *testing.T, stub *testkit.Stub) {
	t.Helper()
	stub.On("x/web-interface/view?aid=626497566", fixture(t, "view-single.json"))
	stub.On("x/web-interface/view?aid=246993280", fixture(t, "view-multi.json"))
	stub.On("pgc/view/web/season?ep_id=317690", fixture(t, "pgc-season.json"))
	stub.On("pgc/view/web/season?season_id=33073", fixture(t, "pgc-season.json"))
	stub.On("x/player/wbi/playurl?", fixture(t, "playurl-web-dash.json"))
	stub.On("pgc/player/web/v2/playurl?", fixture(t, "pgc-playurl.json"))
	// A plain av page that is not licensed content stays where it is.
	stub.On("www.bilibili.com/video/av", status(200))
}
