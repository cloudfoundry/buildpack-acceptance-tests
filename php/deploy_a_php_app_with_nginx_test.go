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

	Context("deploying a basic PHP app using Nginx as the webserver", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(testdata, "with_nginx"))
			app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
			app.SetEnv("LOG_LEVEL", os.Getenv("LOG_LEVEL"))
			PushAppAndConfirm(app)
		})

		It("succeeds", func() {
			By("shows the current buildpack version for useful info")
			Expect(app.Stdout.String()).To(ContainSubstring("-----> Php Buildpack version " + packagedBuildpack.Version))

			By("installs nginx, the request web server")
			Expect(app.Stdout.String()).To(MatchRegexp(`Nginx Server.*1\.\d+\.\d+.*Contributing.* to layer`))

			By("the root endpoint renders a dynamic message")
			Expect(app.GetBody("/")).To(ContainSubstring("PHP Version"))
		})
	})
})
