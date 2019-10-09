package integration_test

import (
	"encoding/json"
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
	bpDir             string
	testdata          string
	buildpackVersion  string
	packagedBuildpack cutlass.VersionedBuildpackPackage
)

func init() {
	flag.StringVar(&buildpackVersion, "version", "", "version to use (builds if empty)")
	flag.BoolVar(&cutlass.Cached, "cutlass.cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	currentDir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	testdata = filepath.Join(currentDir, "testdata")
	bpDir = os.Getenv("BUILDPACK_DIR")
	if bpDir == "" {
		Fail("setting $BUILDPACK_DIR is required")
	}
	// Run once
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
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.RemovePackagedBuildpack(packagedBuildpack)).To(Succeed())
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 60*time.Second).Should(Equal([]string{"RUNNING"}))
	Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
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
	Context("with an uncached buildpack", func() {
		BeforeEach(func() {
			if cutlass.Cached {
				Skip("Running cached tests")
			}
		})

		It("uses a proxy during staging if present", func() {
			bpFile := filepath.Join(bpDir, buildpackVersion+cutlass.RandStringRunes(6)+"tmp")
			cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
			err := cmd.Run()
			Expect(err).To(BeNil())
			defer os.Remove(bpFile)

			Expect(cutlass.EnsureUsesProxy(fixturePath, bpFile)).To(Succeed())
		})
	})
}

func AssertNoInternetTraffic(fixturePath string) {
	if !cutlass.Cached {
		Skip("Running uncached tests")
	}

	randPostFix := cutlass.RandStringRunes(8)
	bpFile := filepath.Join(bpDir, buildpackVersion+"tmp"+randPostFix)
	cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
	err := cmd.Run()
	Expect(err).To(BeNil())
	defer os.Remove(bpFile)

	traffic, built, _, err := cutlass.InternetTraffic(
		bpDir,
		fixturePath,
		bpFile,
		[]string{},
	)
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
