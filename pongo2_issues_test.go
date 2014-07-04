package pongo2

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.

func TestIssues(t *testing.T) { TestingT(t) }

type IssueTestSuite struct{}

var _ = Suite(&IssueTestSuite{})

func (s *TestSuite) TestIssues(c *C) {
	// Add a test for any issue
	c.Check(42, Equals, 42)
}
