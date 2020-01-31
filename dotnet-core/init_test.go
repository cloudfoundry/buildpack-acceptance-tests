package integration_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/agouti"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	packagedBuildpack cutlass.VersionedBuildpackPackage
	agoutiDriver      *agouti.WebDriver
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.StringVar(&packagedBuildpack.File, "buildpack", "", "path the the buildpack")
	flag.StringVar(&packagedBuildpack.Version, "buildpack-version", "", "version to use (builds if empty)")

	flag.BoolVar(&cutlass.Cached, "cutlass.cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "512M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return []byte{}
}, func(data []byte) {
	Expect(cutlass.CopyCfHome()).To(Succeed())

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter

	err := cutlass.CreateOrUpdateBuildpack("dotnet_core", packagedBuildpack.File, os.Getenv("CF_STACK"))
	Expect(err).NotTo(HaveOccurred())

	agoutiDriver = agouti.ChromeDriver(agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}))
	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
	err := agoutiDriver.Stop()
	Expect(err).To(BeNil())
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func GetManifest(path string) (libbuildpack.Manifest, error) {
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return libbuildpack.Manifest{}, err
	}

	for _, header := range zipReader.File {
		if filepath.Clean(header.Name) == "manifest.yml" {
			file, err := header.Open()
			if err != nil {
				return libbuildpack.Manifest{}, err
			}

			var m libbuildpack.Manifest
			err = yaml.NewDecoder(file).Decode(&m)
			if err != nil {
				return libbuildpack.Manifest{}, err
			}

			return m, nil

		}
	}

	return libbuildpack.Manifest{}, errors.New("failed to find manifest.yml in buildpack")
}

func Fixtures(names ...string) string {
	currentDir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	names = append([]string{currentDir, "testdata"}, names...)
	return filepath.Join(names...)
}

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
	Expect(app.ConfirmBuildpack(packagedBuildpack.Version)).To(Succeed())
}

func Restart(app *cutlass.App) {
	Expect(app.Restart()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
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

func SkipUnlessStack(requiredStack string) {
	currentStack := os.Getenv("CF_STACK")
	if currentStack != requiredStack {
		Skip(fmt.Sprintf("Skipping because the stack \"%s\" is not supported", currentStack))
	}
}

func DestroyApp(app *cutlass.App) *cutlass.App {
	if app != nil {
		app.Destroy()
	}
	return nil
}

func DefaultVersion(name string) string {
	m, err := GetManifest(packagedBuildpack.File)
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

			traffic, built, _, err := cutlass.InternetTraffic(
				Fixtures(fixtureName),
				packagedBuildpack.File,
				[]string{"HTTP_PROXY=" + proxy.URL, "HTTPS_PROXY=" + proxy.URL},
			)
			Expect(err).To(BeNil())
			Expect(built).To(BeTrue())

			destUrl, err := url.Parse(proxy.URL)
			Expect(err).To(BeNil())

			Expect(cutlass.UniqueDestination(
				traffic, fmt.Sprintf("%s.%s", destUrl.Hostname(), destUrl.Port()),
			)).To(BeNil())
		})
	})
}

func AssertNoInternetTraffic(fixturePath string) {
	SkipUnlessCached()

	traffic, _, _, err := cutlass.InternetTraffic(
		fixturePath,
		packagedBuildpack.File,
		[]string{},
	)
	Expect(err).To(BeNil())
	Expect(traffic).To(BeEmpty())
}

func GetLatestDepVersion(dep, constraint string) string {
	manifest, err := GetManifest(packagedBuildpack.File)
	Expect(err).ToNot(HaveOccurred())

	fmt.Printf("manifest -> %#v\n", manifest)

	deps := manifest.AllDependencyVersions(dep)
	runtimeVersion, err := libbuildpack.FindMatchingVersion(constraint, deps)
	Expect(err).ToNot(HaveOccurred())

	return runtimeVersion
}

func ReplaceFileTemplate(pathToFixture, file, templateVar, replaceVal string) *cutlass.App {
	dir, err := cutlass.CopyFixture(pathToFixture)
	Expect(err).ToNot(HaveOccurred())

	data, err := ioutil.ReadFile(filepath.Join(dir, file))
	Expect(err).ToNot(HaveOccurred())
	data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", templateVar)), []byte(replaceVal), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, file), data, 0644)).To(Succeed())

	return cutlass.New(dir)
}

func PrintFailureLogs(appName string) error {
	if !CurrentGinkgoTestDescription().Failed {
		return nil
	}
	command := exec.Command("cf", "logs", appName, "--recent")
	command.Stdout = GinkgoWriter
	command.Stderr = GinkgoWriter
	return command.Run()
}
