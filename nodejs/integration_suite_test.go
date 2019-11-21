package integration_test

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	testdata          string
	packagedBuildpack cutlass.VersionedBuildpackPackage
)

var _ = func() bool {
	testing.Init()
	return true
}()

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
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	currentDir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	testdata = filepath.Join(currentDir, "testdata")

	Expect(cutlass.CopyCfHome()).To(Succeed())
	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter

	err = cutlass.CreateOrUpdateBuildpack("nodejs", packagedBuildpack.File, os.Getenv("CF_STACK"))
	Expect(err).NotTo(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 60*time.Second).Should(Equal([]string{"RUNNING"}))
	Expect(app.ConfirmBuildpack(packagedBuildpack.Version)).To(Succeed())
}

func DestroyApp(app *cutlass.App) *cutlass.App {
	if app != nil {
		app.Destroy()
	}
	return nil
}

func ApiHasTask() bool {
	supported, err := cutlass.ApiGreaterThan("2.75.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func ApiHasMultiBuildpack() bool {
	supported, err := cutlass.ApiGreaterThan("2.90.0")
	Expect(err).NotTo(HaveOccurred(), "the targeted CF does not support multiple buildpacks")
	return supported
}

func ApiSupportsSymlinks() bool {
	supported, err := cutlass.ApiGreaterThan("2.103.0")
	Expect(err).NotTo(HaveOccurred(), "the targeted CF does not support symlinks")
	return supported
}

func ApiHasStackAssociation() bool {
	supported, err := cutlass.ApiGreaterThan("2.113.0")
	Expect(err).NotTo(HaveOccurred(), "the targeted CF does not support stack association")
	return supported
}

func AssertUsesProxyDuringStagingIfPresent(fixturePath string) {
	if cutlass.Cached {
		Skip("Running cached tests")
	}

	Expect(cutlass.EnsureUsesProxy(fixturePath, packagedBuildpack.File)).To(Succeed())
}

func AssertNoInternetTraffic(fixturePath string) {
	if !cutlass.Cached {
		Skip("Running uncached tests")
	}

	traffic, built, _, err := cutlass.InternetTraffic(fixturePath, packagedBuildpack.File, []string{})
	Expect(err).To(BeNil())
	Expect(built).To(BeTrue())
	Expect(traffic).To(BeEmpty())
}

func RunCF(args ...string) error {
	command := exec.Command("cf", args...)
	command.Stdout = GinkgoWriter
	command.Stderr = GinkgoWriter
	return command.Run()
}
