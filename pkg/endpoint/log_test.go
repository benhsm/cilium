// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

//go:build integration_tests

package endpoint

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"

	"github.com/cilium/cilium/pkg/ipcache"
	"github.com/cilium/cilium/pkg/option"
	"github.com/cilium/cilium/pkg/policy"
	testidentity "github.com/cilium/cilium/pkg/testutils/identity"
)

func (s *EndpointSuite) TestPolicyLog(c *C) {
	do := &DummyOwner{repo: policy.NewPolicyRepository(nil, nil, nil)}
	ep := NewEndpointWithState(do, do, ipcache.NewIPCache(nil), nil, testidentity.NewMockIdentityAllocator(nil), 12345, StateReady)

	// Initially nil
	policyLogger := ep.getPolicyLogger()
	c.Assert(policyLogger, IsNil)

	// Enable DebugPolicy option
	ep.Options.SetValidated(option.DebugPolicy, option.OptionEnabled)
	c.Assert(ep.Options.IsEnabled(option.DebugPolicy), Equals, true)
	ep.UpdateLogger(nil)
	policyLogger = ep.getPolicyLogger()
	c.Assert(policyLogger, Not(IsNil))
	defer func() {
		// remote created log file when we are done.
		err := os.Remove(filepath.Join(option.Config.StateDir, "endpoint-policy.log"))
		c.Assert(err, IsNil)
	}()

	// Test logging, policyLogger must not be nil
	policyLogger.Info("testing policy logging")

	// Test logging with integrated nil check, no fields
	ep.PolicyDebug(nil, "testing PolicyDebug")
	ep.PolicyDebug(logrus.Fields{"testField": "Test Value"}, "PolicyDebug with fields")

	// Disable option
	ep.Options.SetValidated(option.DebugPolicy, option.OptionDisabled)
	c.Assert(ep.Options.IsEnabled(option.DebugPolicy), Equals, false)
	ep.UpdateLogger(nil)
	policyLogger = ep.getPolicyLogger()
	c.Assert(policyLogger, IsNil)

	// Verify file exists and contains the logged message
	buf, err := os.ReadFile(filepath.Join(option.Config.StateDir, "endpoint-policy.log"))
	c.Assert(err, IsNil)
	c.Assert(bytes.Contains(buf, []byte("testing policy logging")), Equals, true)
	c.Assert(bytes.Contains(buf, []byte("testing PolicyDebug")), Equals, true)
	c.Assert(bytes.Contains(buf, []byte("Test Value")), Equals, true)
}
