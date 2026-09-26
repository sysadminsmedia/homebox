package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// freePort reserves a port, then releases it so the code under test can bind it.
func freePort(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

// isActivationAddress is the gate that keeps anyhttp — and therefore its
// implicit bind of ":http" (port 80) for an empty address — off the plain TCP
// path entirely. These cases mirror anyhttp's own parseAddress, so a drift in
// either direction shows up here. See issue #1656.
func TestIsActivationAddress(t *testing.T) {
	t.Parallel()

	activation := []string{
		"unix?path=/tmp/homebox.sock",
		"unix?path=/tmp/homebox.sock&mode=660",
		"sysd",
		"sysd?name=homebox",
	}
	for _, addr := range activation {
		assert.True(t, isActivationAddress(addr), "%q should be an activation address", addr)
	}

	plain := []string{
		"", // the default: anyhttp would map this to ":http" and bind port 80
		"0.0.0.0",
		"127.0.0.1",
		"localhost",
		"0.0.0.0:7745",
		"127.0.0.1:7745",
	}
	for _, addr := range plain {
		assert.False(t, isActivationAddress(addr), "%q should be handled as plain TCP", addr)
	}
}

// anyhttp opens a TCP socket for any address that isn't a unix socket or a
// systemd FD, including the ":http" (port 80) default it substitutes for an
// empty Web.Host. We never serve that socket, so it must not end up bound.
// See issue #1656.
func TestActivationListenerReleasesPlainTCPSocket(t *testing.T) {
	addr := freePort(t)

	listener, err := activationListener(addr)
	require.NoError(t, err)
	assert.Nil(t, listener, "a plain TCP address must not yield a listener to Serve")

	// The real assertion: the port anyhttp bound has to be free again.
	reclaimed, err := net.Listen("tcp", addr)
	require.NoError(t, err, "anyhttp's unused TCP listener was left bound (issue #1656)")
	require.NoError(t, reclaimed.Close())
}

// The default configuration leaves Web.Host empty, which is the case that
// reached out and bound port 80.
func TestActivationListenerEmptyHostYieldsNoListener(t *testing.T) {
	listener, err := activationListener("")
	require.NoError(t, err)
	assert.Nil(t, listener,
		"an empty Web.Host must fall through to ListenAndServe, not serve anyhttp's :80 socket")
}

// A unix socket address is a genuine activation mode and must still produce a
// listener for the server to Serve.
func TestActivationListenerReturnsUnixSocket(t *testing.T) {
	sock := t.TempDir() + "/homebox.sock"

	listener, err := activationListener("unix?path=" + sock + "&remove_existing=true")
	require.NoError(t, err)
	require.NotNil(t, listener, "a unix socket address must yield a listener to Serve")
	t.Cleanup(func() { _ = listener.Close() })

	assert.Equal(t, "unix", listener.Addr().Network())
}
