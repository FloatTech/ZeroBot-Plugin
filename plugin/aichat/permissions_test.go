package aichat

import (
	"testing"

	goba "github.com/fumiama/go-onebot-agent"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func TestAgentRoleUsesCurrentGroupAndConfiguredOwner(t *testing.T) {
	old := zero.BotConfig.SuperUsers
	zero.BotConfig.SuperUsers = []int64{123}
	t.Cleanup(func() { zero.BotConfig.SuperUsers = old })
	for _, c := range []struct {
		name     string
		uid, gid int64
		sender   *zero.User
		want     goba.PermRole
	}{
		{"owner private", 123, 0, &zero.User{Role: "member"}, goba.PermRoleOwner},
		{"owner group member", 123, 101, &zero.User{Role: "member"}, goba.PermRoleOwner},
		{"group owner", 456, 101, &zero.User{Role: "owner"}, goba.PermRoleAdmin},
		{"group admin", 456, 101, &zero.User{Role: "admin"}, goba.PermRoleAdmin},
		{"group member", 456, 101, &zero.User{Role: "member"}, goba.PermRoleUser},
		{"unknown group role", 456, 101, &zero.User{Role: "unexpected"}, goba.PermRoleUser},
		{"missing sender", 456, 101, nil, goba.PermRoleUser},
		{"ordinary private", 456, 0, &zero.User{Role: "member"}, goba.PermRoleUser},
		{"private role cannot grant admin", 456, 0, &zero.User{Role: "admin"}, goba.PermRoleUser},
		{"private role cannot grant owner", 456, 0, &zero.User{Role: "owner"}, goba.PermRoleUser},
	} {
		t.Run(c.name, func(t *testing.T) {
			ctx := &zero.Ctx{Event: &zero.Event{UserID: c.uid, GroupID: c.gid, Sender: c.sender}}
			if got := agentRole(ctx); got != c.want {
				t.Fatalf("got %s, want %s", got, c.want)
			}
		})
	}
}
