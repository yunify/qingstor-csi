package cache_test

import (
	"testing"

	"flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var hasCli bool

const (
	UnsupportCli = "Unsupport NeonSAN CLI"
)

func init() {
	flag.BoolVar(&hasCli, "hasCli", false, "current environment support NeonSAN CLI")
}

func TestCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}
