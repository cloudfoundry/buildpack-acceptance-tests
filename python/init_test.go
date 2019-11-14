package integration_test

import (
	"flag"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var packagedBuildpack cutlass.VersionedBuildpackPackage

func init() {
	flag.StringVar(&packagedBuildpack.File, "buildpack", "", "path to a packaged buildpack")
	flag.StringVar(&packagedBuildpack.Version, "buildpack-version", "", "version of the packaged buildpack")
	flag.BoolVar(&cutlass.Cached, "cutlass.cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Python Integration Suite")
}
