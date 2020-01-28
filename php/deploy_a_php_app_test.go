package integration_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF PHP Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(testdata, "php_app"))
		app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
		app.SetEnv("BP_DEBUG", os.Getenv("LOG_LEVEL"))
	})

	It("deploying a basic PHP app", func() {
		PushAppAndConfirm(app)

		By("installs our hard-coded default version of PHP")
		Expect(app.Stdout.String()).To(MatchRegexp(`PHP.*7\.2\.\d+.*Contributing.* to layer`))

		By("does not return the version of PHP in the response headers")
		body, headers, err := app.Get("/", map[string]string{})
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(ContainSubstring("PHP Version"))
		Expect(headers).ToNot(HaveKey("X-Powered-By"))

		if cutlass.Cached {
			By("downloads the binaries directly from the buildpack")
			// use of RegEx is required because of weird color/terminal control characters from libcfbuildpack
			Expect(app.Stdout.String()).To(MatchRegexp(`Reusing.*cached download from buildpack`))
			AssertNoInternetTraffic("testdata/php_app")
		}
	})
})
