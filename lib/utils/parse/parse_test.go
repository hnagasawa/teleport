/*
Copyright 2017-2020 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package parse

import (
	"fmt"
	"testing"

	"github.com/gravitational/teleport/lib/utils"

	"github.com/gravitational/trace"
	"gopkg.in/check.v1"
)

func TestParse(t *testing.T) { check.TestingT(t) }

type ParseSuite struct{}

var _ = check.Suite(&ParseSuite{})
var _ = fmt.Printf

func (s *ParseSuite) SetUpSuite(c *check.C) {
	utils.InitLoggerForTests()
}
func (s *ParseSuite) TearDownSuite(c *check.C) {}
func (s *ParseSuite) SetUpTest(c *check.C)     {}
func (s *ParseSuite) TearDownTest(c *check.C)  {}

// TestRoleVariable tests variable parsing
func (s *ParseSuite) TestRoleVariable(c *check.C) {
	var tests = []struct {
		title string
		in    string
		err   error
		out   Expression
	}{
		{
			title: "no curly bracket prefix",
			in:    "external.foo}}",
			err:   trace.BadParameter(""),
		},
		{
			title: "invalid syntax",
			in:    `{{external.foo("bar")`,
			err:   trace.BadParameter(""),
		},
		{
			title: "invalid variable syntax",
			in:    "{{internal.}}",
			err:   trace.BadParameter(""),
		},
		{
			title: "invalid dot syntax",
			in:    "{{external..foo}}",
			err:   trace.BadParameter(""),
		},
		{
			title: "empty variable",
			in:    "{{}}",
			err:   trace.BadParameter(""),
		},
		{
			title: "no curly bracket suffix",
			in:    "{{internal.foo",
			err:   trace.BadParameter(""),
		},
		{
			title: "too many levels of nesting in the variable",
			in:    "{{internal.foo.bar}}",
			err:   trace.BadParameter(""),
		},
		{
			title: "valid with brackets",
			in:    `{{internal["foo"]}}`,
			out:   Expression{namespace: "internal", variable: "foo"},
		},
		{
			title: "external with no brackets",
			in:    "{{external.foo}}",
			out:   Expression{namespace: "external", variable: "foo"},
		},
		{
			title: "internal with no brackets",
			in:    "{{internal.bar}}",
			out:   Expression{namespace: "internal", variable: "bar"},
		},
		{
			title: "internal with spaces removed",
			in:    "  {{  internal.bar  }}  ",
			out:   Expression{namespace: "internal", variable: "bar"},
		},
		{
			title: "variable with prefix and suffix",
			in:    "  hello,  {{  internal.bar  }}  there! ",
			out:   Expression{prefix: "hello,  ", namespace: "internal", variable: "bar", suffix: "  there!"},
		},
	}

	for i, tt := range tests {
		comment := check.Commentf("Test(%v) %q", i, tt.title)

		variable, err := RoleVariable(tt.in)
		if tt.err != nil {
			c.Assert(err, check.FitsTypeOf, tt.err, comment)
			continue
		}
		c.Assert(variable, check.DeepEquals, &tt.out, comment)
	}
}

// TestInterpolate tests variable interpolation
func (s *ParseSuite) TestInterpolate(c *check.C) {
	type result struct {
		values []string
		found  bool
	}
	var tests = []struct {
		title  string
		in     Expression
		traits map[string][]string
		res    result
	}{
		{
			title:  "mapped traits",
			in:     Expression{variable: "foo"},
			traits: map[string][]string{"foo": []string{"a", "b"}, "bar": []string{"c"}},
			res:    result{found: true, values: []string{"a", "b"}},
		},
		{
			title:  "missed traits",
			in:     Expression{variable: "baz"},
			traits: map[string][]string{"foo": []string{"a", "b"}, "bar": []string{"c"}},
			res:    result{found: false, values: []string{}},
		},
		{
			title:  "traits with prefix and suffix",
			in:     Expression{prefix: "IAM#", variable: "foo", suffix: ";"},
			traits: map[string][]string{"foo": []string{"a", "b"}, "bar": []string{"c"}},
			res:    result{found: true, values: []string{"IAM#a;", "IAM#b;"}},
		},
	}

	for i, tt := range tests {
		comment := check.Commentf("Test(%v) %q", i, tt.title)

		values, found := tt.in.Interpolate(tt.traits)
		if !tt.res.found {
			c.Assert(found, check.Equals, false, comment)
			c.Assert(values, check.HasLen, 0)
			continue
		}

		c.Assert(values, check.DeepEquals, tt.res.values, comment)
	}
}
