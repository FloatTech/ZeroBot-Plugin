package localadmin

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type policyCase struct {
	name, table, event, where string
	prepare, apply            func(*ctrl.Control[*zero.Ctx]) error
	presentBefore             bool
}

func policyCases() []policyCase {
	return []policyCase{
		{"response", "__resp", "INSERT", "WHERE gid=101", nil, func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.Response(101) }, false},
		{"silence", "__resp", "DELETE", "WHERE gid=101", func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.Response(101) }, func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.Silence(101) }, true},
		{"global-response", "__resp", "INSERT", "WHERE gid=0", nil, func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.Response(0) }, false},
		{"global-silence", "__resp", "DELETE", "WHERE gid=0", func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.Response(0) }, func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.Silence(0) }, true},
		{"block", "__block", "INSERT", "WHERE uid=123", nil, func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.DoBlock(123) }, false},
		{"unblock", "__block", "DELETE", "WHERE uid=123", func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.DoBlock(123) }, func(c *ctrl.Control[*zero.Ctx]) error { return c.Manager.DoUnblock(123) }, true},
		{"ban", "test-pluginban", "INSERT", "WHERE uid=123 AND gid=101", nil, func(c *ctrl.Control[*zero.Ctx]) error { return c.Ban(123, 101) }, false},
		{"permit", "test-pluginban", "DELETE", "WHERE uid=123 AND gid=101", func(c *ctrl.Control[*zero.Ctx]) error { return c.Ban(123, 101) }, func(c *ctrl.Control[*zero.Ctx]) error { return c.Permit(123, 101) }, true},
	}
}

func policyCaches(c *ctrl.Control[*zero.Ctx]) [4]bool {
	return [4]bool{c.Manager.CanResponse(101), c.Manager.CanResponse(0), c.Manager.IsBlocked(123), c.IsBannedIn(123, 101)}
}

func policyBackupEntries(t *testing.T, directory string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func TestPolicySnapshotsIncludePreWriteWALStateAndSkipNoOps(t *testing.T) {
	for _, tc := range policyCases() {
		t.Run(tc.name, func(t *testing.T) {
			c, dbPath := pluginFixture(t, false)
			if _, err := c.Manager.D.Exec("PRAGMA journal_mode=WAL"); err != nil {
				t.Fatal(err)
			}
			if tc.prepare != nil {
				if err := tc.prepare(c); err != nil {
					t.Fatal(err)
				}
			}
			// A committed marker still in WAL must be included in every snapshot.
			if err := c.Manager.D.Insert(c.Service, &ctrl.GroupConfig{GroupID: 202, Disable: 34}); err != nil {
				t.Fatal(err)
			}
			backups := filepath.Join(t.TempDir(), "backups")
			NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, backups)
			if err := tc.apply(c); err != nil {
				t.Fatal(err)
			}
			entries := policyBackupEntries(t, backups)
			if len(entries) != 1 {
				t.Fatalf("got %d snapshots", len(entries))
			}
			copyDB := sql.New(filepath.Join(backups, entries[0].Name(), "plugins.sqlite"))
			if err := copyDB.Open(time.Hour); err != nil {
				t.Fatal(err)
			}
			defer copyDB.Close()
			if copyDB.CanFind(tc.table, tc.where) != tc.presentBefore {
				t.Fatal("snapshot contains post-write policy")
			}
			var marker ctrl.GroupConfig
			if err := copyDB.Find(c.Service, &marker, "WHERE gid=202"); err != nil || marker.Disable != 34 {
				t.Fatal("snapshot omitted WAL marker")
			}
			if c.Manager.D.CanFind(tc.table, tc.where) == tc.presentBefore {
				t.Fatal("policy was not committed")
			}
			before := policyCaches(c)
			m := ctrl.NewManager[*zero.Ctx](dbPath)
			reloaded := m.NewControl(c.Service, &ctrl.Options[*zero.Ctx]{})
			if policyCaches(reloaded) != before {
				t.Fatal("policy did not survive a fresh manager")
			}
			_ = m.D.Close()
			err := tc.apply(c)
			if err != nil && !errors.Is(err, ctrl.ErrAlreadyResponding) && !errors.Is(err, ctrl.ErrAlreadySilent) {
				t.Fatal(err)
			}
			if len(policyBackupEntries(t, backups)) != 1 {
				t.Fatal("duplicate policy created a snapshot")
			}
		})
	}
}

func TestPolicyBackupAndSQLFailurePreserveDiskAndCaches(t *testing.T) {
	for _, mode := range []string{"backup", "sql", "read"} {
		for _, tc := range policyCases() {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				c, _ := pluginFixture(t, false)
				if tc.prepare != nil {
					if err := tc.prepare(c); err != nil {
						t.Fatal(err)
					}
				}
				before := policyCaches(c)
				backups := filepath.Join(t.TempDir(), "backups")
				NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, backups)
				switch mode {
				case "backup":
					if err := os.WriteFile(backups, []byte("blocked"), 0600); err != nil {
						t.Fatal(err)
					}
				case "sql":
					if _, err := c.Manager.D.Exec("CREATE TRIGGER reject_policy BEFORE " + tc.event + " ON \"" + tc.table + "\" BEGIN SELECT RAISE(ABORT,'synthetic-secret'); END"); err != nil {
						t.Fatal(err)
					}
				case "read":
					// Force the persisted-state lookup to fail without removing cached state.
					if _, err := c.Manager.D.Exec("ALTER TABLE \"" + tc.table + "\" RENAME TO hidden_policy"); err != nil {
						t.Fatal(err)
					}
				}
				if err := tc.apply(c); err == nil {
					t.Fatal("write unexpectedly succeeded")
				}
				if policyCaches(c) != before {
					t.Fatal("failed write changed cache")
				}
				table := tc.table
				if mode == "read" {
					table = "hidden_policy"
				}
				if c.Manager.D.CanFind(table, tc.where) != tc.presentBefore {
					t.Fatal("failed write changed disk")
				}
				if mode == "sql" && len(policyBackupEntries(t, backups)) != 1 {
					t.Fatal("write failure lost recovery snapshot")
				}
				if mode == "read" && len(policyBackupEntries(t, backups)) != 0 {
					t.Fatal("read failure invoked backup")
				}
			})
		}
	}
}

func TestPolicyGlobalFallbackAndConcurrentSnapshots(t *testing.T) {
	c, _ := pluginFixture(t, false)
	backups := filepath.Join(t.TempDir(), "backups")
	NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, backups)
	if err := c.Manager.Response(0); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(c.Manager.Response(101), ctrl.ErrAlreadyResponding) {
		t.Fatal("global response precedence changed")
	}
	if err := c.Manager.Silence(101); err != nil {
		t.Fatal(err)
	}
	if !c.Manager.CanResponse(101) || len(policyBackupEntries(t, backups)) != 1 {
		t.Fatal("absent local row changed global response or backed up")
	}
	if err := c.Ban(123, 0); err != nil {
		t.Fatal(err)
	}
	if err := c.Permit(123, 101); err != nil {
		t.Fatal(err)
	}
	if !c.IsBannedIn(123, 101) || len(policyBackupEntries(t, backups)) != 2 {
		t.Fatal("local permit changed global ban or made snapshot")
	}
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- c.Manager.DoBlock(123) }()
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !c.Manager.IsBlocked(123) || len(policyBackupEntries(t, backups)) != 3 {
		t.Fatal("concurrent duplicate writes did not share a single snapshot")
	}
}
