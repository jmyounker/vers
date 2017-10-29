package main

import (
	"testing"
)


func TestParseSvnInfo(t *testing.T) {
	svnOut := "Path: foo.c\n" +
		"Name: foo.c\n" +
		"URL: http://svn.red-bean.com/repos/test/foo.c\n" +
		"Repository Root: http://svn.red-bean.com/repos/test\n" +
		"Repository UUID: 5e7d134a-54fb-0310-bd04-b611643e5c25\n" +
		"Revision: 4417\n" +
		"Node Kind: file\n" +
		"Schedule: normal\n" +
		"Last Changed Author: sally\n" +
		"Last Changed Rev: 20\n" +
		"Last Changed Date: 2003-01-13 16:43:13 -0600 (Mon, 13 Jan 2003)\n" +
		"Text Last Updated: 2003-01-16 21:18:16 -0600 (Thu, 16 Jan 2003)\n" +
		"Properties Last Updated: 2003-01-13 21:50:19 -0600 (Mon, 13 Jan 2003)\n" +
		"Checksum: d6aeb60b0662ccceb6bce4bac344cb66\n"
	m, err := ParseSvnInfo(svnOut)
	failWhenErr(t, err)
	v, ok := m["Repository Root"]
	failWhen(t, !ok)
	failWhen(t, v != "http://svn.red-bean.com/repos/test" )
}

func TestParseRevisionFromXmlLog(t *testing.T) {
	svnOut := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<log>\n" +
		"<logentry\n" +
		"revision=\"1\">\n" +
		"<author>jeff</author>\n" +
		"<date>2017-10-29T11:45:01.882457Z</date>\n" +
		"<msg>Add a file." +
		"</msg>\n" +
		"</logentry>\n" +
		"</log>\n"
	m, err := ParseRevisionFromXmlLog(svnOut)
	failWhenErr(t, err)
	failWhen(t, m != "1")
}

func TestParseRevisionFromXmlLogWithEmptyRepo(t *testing.T) {
	svnOut := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<log>\n" +
		"</log>\n"
	m, err := ParseRevisionFromXmlLog(svnOut)
	failWhenErr(t, err)
	failWhen(t, m != "0")
}

func TestParseBranch(t *testing.T) {
	var cases = []struct{
		Url string
		Branch string
	}{
		{ "http://svn.red-bean.com/repos/trunk", "trunk" },
		{ "http://svn.red-bean.com/repos/trunk/foo", "trunk" },
		{ "http://svn.red-bean.com/repos/branches/foo", "foo" },
		{ "http://svn.red-bean.com/repos/branches/foo/bar", "foo" },
		{ "http://svn.red-bean.com/repos/tags/foo", "foo" },
		{ "http://svn.red-bean.com/repos/tags/foo/bar", "foo" },
	}
	for _, tc := range(cases) {
		b, err := ParseBranchFromSvnUrl(tc.Url)
		failWhenErr(t, err)
		failWhen(t, b != tc.Branch)
	}
}
