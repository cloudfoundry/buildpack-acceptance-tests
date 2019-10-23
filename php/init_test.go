package integration_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	bpDir             string
	testdata          string
	buildpackVersion  string
	packagedBuildpack cutlass.VersionedBuildpackPackage
)

func init() {
	flag.StringVar(&buildpackVersion, "version", "", "version to use (builds if empty)")
	flag.BoolVar(&cutlass.Cached, "cutlass.cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "384M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	currentDir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	testdata = filepath.Join(currentDir, "testdata")
	bpDir = os.Getenv("BUILDPACK_DIR")
	if bpDir == "" {
		Fail("setting $BUILDPACK_DIR is required")
	}

	if os.Getenv("GIT_TOKEN") != "" {
		os.Setenv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("GIT_TOKEN"))
	}

	Expect(os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN")).ToNot(BeEmpty(), "Please set COMPOSER_GITHUB_OAUTH_TOKEN") // Required for some tests

	if buildpackVersion == "" {
		Expect(os.Chdir(bpDir)).To(Succeed())

		packagedBuildpack, err := cutlass.PackageShimmedBuildpack(bpDir, os.Getenv("CF_STACK"))
		Expect(err).NotTo(HaveOccurred(), "failed to package buildpack")

		Expect(os.Chdir(currentDir)).To(Succeed())

		data, err := json.Marshal(packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		return data
	}

	return nil
}, func(data []byte) {
	// Run on all nodes
	var err error
	if len(data) > 0 {
		err = json.Unmarshal(data, &packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		buildpackVersion = packagedBuildpack.Version
	}

	Expect(cutlass.CopyCfHome()).To(Succeed())

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter

	SetDefaultEventuallyTimeout(10 * time.Second)
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	cutlass.RemovePackagedBuildpack(packagedBuildpack)
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func ConfirmRunning(app *cutlass.App) {
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	ConfirmRunning(app)
	Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
}

func Restart(app *cutlass.App) {
	Expect(app.Restart()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func Fixtures(names ...string) string {
	names = append([]string{testdata}, names...)
	return filepath.Join(names...)
}

func ApiGreaterThan(version string) bool {
	apiVersionString, err := cutlass.ApiVersion()
	Expect(err).To(BeNil())
	apiVersion, err := semver.Make(apiVersionString)
	Expect(err).To(BeNil())
	reqVersion, err := semver.ParseRange(">= " + version)
	Expect(err).To(BeNil())
	return reqVersion(apiVersion)
}

func ApiHasTask() bool {
	supported, err := cutlass.ApiGreaterThan("2.75.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func ApiHasMultiBuildpack() bool {
	supported, err := cutlass.ApiGreaterThan("2.90.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func ApiHasStackAssociation() bool {
	supported, err := cutlass.ApiGreaterThan("2.113.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func SkipUnlessUncached() {
	if cutlass.Cached {
		Skip("Running cached tests")
	}
}

func SkipUnlessCached() {
	if !cutlass.Cached {
		Skip("Running uncached tests")
	}
}

func DestroyApp(app *cutlass.App) *cutlass.App {
	if app != nil {
		app.Destroy()
	}
	return nil
}

func DefaultVersion(name string) string {
	m := &libbuildpack.Manifest{}
	err := (&libbuildpack.YAML{}).Load(filepath.Join(bpDir, "manifest.yml"), m)
	Expect(err).ToNot(HaveOccurred())
	dep, err := m.DefaultVersion(name)
	Expect(err).ToNot(HaveOccurred())
	Expect(dep.Version).ToNot(Equal(""))
	return dep.Version
}

func AssertUsesProxyDuringStagingIfPresent(fixtureName string) {
	Context("with an uncached buildpack", func() {
		BeforeEach(SkipUnlessUncached)

		It("uses a proxy during staging if present", func() {
			proxy, err := cutlass.NewProxy()
			Expect(err).To(BeNil())
			defer proxy.Close()

			bpFile := filepath.Join(bpDir, buildpackVersion+"tmp")
			cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
			err = cmd.Run()
			Expect(err).To(BeNil())
			defer os.Remove(bpFile)

			traffic, _, _, err := cutlass.InternetTraffic(
				bpDir,
				Fixtures(fixtureName),
				bpFile,
				[]string{"HTTP_PROXY=" + proxy.URL, "HTTPS_PROXY=" + proxy.URL},
			)
			Expect(err).To(BeNil())
			// Expect(built).To(BeTrue())

			destUrl, err := url.Parse(proxy.URL)
			Expect(err).To(BeNil())

			Expect(cutlass.UniqueDestination(
				traffic, fmt.Sprintf("%s.%s", destUrl.Hostname(), destUrl.Port()),
			)).To(BeNil())
		})
	})
}

func AssertNoInternetTraffic(fixtureName string) {
	It("has no traffic", func() {
		SkipUnlessCached()

		bpFile := filepath.Join(bpDir, buildpackVersion+"tmp"+cutlass.RandStringRunes(8))
		cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
		err := cmd.Run()
		Expect(err).To(BeNil())
		defer os.Remove(bpFile)

		traffic, _, _, err := cutlass.InternetTraffic(
			bpDir,
			Fixtures(fixtureName),
			bpFile,
			[]string{},
		)
		Expect(err).To(BeNil())
		// Expect(built).To(BeTrue())
		Expect(traffic).To(BeEmpty())
	})
}
