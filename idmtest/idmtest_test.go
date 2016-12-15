// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package idmtest_test

import (
	"time"

	jc "github.com/juju/testing/checkers"
	"golang.org/x/net/context"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"

	"github.com/juju/idmclient"
	"github.com/juju/idmclient/idmtest"
	idmparams "github.com/juju/idmclient/params"
)

type suite struct{}

var _ = gc.Suite(&suite{})

func (*suite) TestDischarge(c *gc.C) {
	ctx := context.TODO()
	srv := idmtest.NewServer()
	srv.AddUser("bob", "somegroup")
	client := srv.Client("bob")

	key, err := bakery.GenerateKey()
	c.Assert(err, gc.IsNil)
	b := bakery.New(bakery.BakeryParams{
		Key:            key,
		Locator:        srv,
		IdentityClient: srv.IDMClient("server-user"),
	})
	m, err := b.Oven.NewMacaroon(
		ctx,
		bakery.LatestVersion,
		time.Now().Add(time.Minute),
		idmclient.IdentityCaveats(srv.URL.String()),
		bakery.LoginOp,
	)
	c.Assert(err, gc.IsNil)

	ms, err := client.DischargeAll(ctx, m)
	c.Assert(err, gc.IsNil)

	// Make sure that the macaroon discharged correctly and that it
	// has the right declared caveats.
	authInfo, err := b.Checker.Auth(ms).Allow(ctx, bakery.LoginOp)
	c.Assert(err, gc.IsNil)
	c.Assert(authInfo.Identity, gc.NotNil)
	ident := authInfo.Identity.(idmclient.Identity)
	c.Assert(ident.Id(), gc.Equals, "bob")
	username, err := ident.Username()
	c.Assert(err, gc.IsNil)
	c.Assert(username, gc.Equals, "bob")
	groups, err := ident.Groups()
	c.Assert(err, gc.IsNil)
	c.Assert(groups, jc.DeepEquals, []string{"somegroup"})
}

func (*suite) TestDischargeDefaultUser(c *gc.C) {
	ctx := context.TODO()
	srv := idmtest.NewServer()
	srv.SetDefaultUser("bob")

	key, err := bakery.GenerateKey()
	c.Assert(err, gc.IsNil)
	b := bakery.New(bakery.BakeryParams{
		Key:            key,
		Locator:        srv,
		IdentityClient: srv.IDMClient("server-user"),
	})
	m, err := b.Oven.NewMacaroon(
		ctx,
		bakery.LatestVersion,
		time.Now().Add(time.Minute),
		idmclient.IdentityCaveats(srv.URL.String()),
		bakery.LoginOp,
	)
	c.Assert(err, gc.IsNil)

	client := httpbakery.NewClient()
	ms, err := client.DischargeAll(ctx, m)
	c.Assert(err, gc.IsNil)

	// Make sure that the macaroon discharged correctly and that it
	// has the right declared caveats.
	authInfo, err := b.Checker.Auth(ms).Allow(ctx, bakery.LoginOp)
	c.Assert(err, gc.IsNil)
	c.Assert(authInfo.Identity, gc.NotNil)
	ident := authInfo.Identity.(idmclient.Identity)
	c.Assert(ident.Id(), gc.Equals, "bob")
	username, err := ident.Username()
	c.Assert(err, gc.IsNil)
	c.Assert(username, gc.Equals, "bob")
	groups, err := ident.Groups()
	c.Assert(err, gc.IsNil)
	c.Assert(groups, gc.HasLen, 0)
}

func (*suite) TestGroups(c *gc.C) {
	srv := idmtest.NewServer()
	srv.AddUser("bob", "beatles", "bobbins")
	srv.AddUser("alice")

	client := srv.IDMClient("bob")
	groups, err := client.UserGroups(context.TODO(), &idmparams.UserGroupsRequest{
		Username: "bob",
	})
	c.Assert(err, gc.IsNil)
	c.Assert(groups, jc.DeepEquals, []string{"beatles", "bobbins"})

	groups, err = client.UserGroups(context.TODO(), &idmparams.UserGroupsRequest{
		Username: "alice",
	})
	c.Assert(err, gc.IsNil)
	c.Assert(groups, gc.HasLen, 0)
}

func (s *suite) TestAddUserWithExistingGroups(c *gc.C) {
	srv := idmtest.NewServer()
	srv.AddUser("alice", "anteaters")
	srv.AddUser("alice")
	srv.AddUser("alice", "goof", "anteaters")

	client := srv.IDMClient("alice")
	groups, err := client.UserGroups(context.TODO(), &idmparams.UserGroupsRequest{
		Username: "alice",
	})
	c.Assert(err, gc.IsNil)
	c.Assert(groups, jc.DeepEquals, []string{"anteaters", "goof"})
}
