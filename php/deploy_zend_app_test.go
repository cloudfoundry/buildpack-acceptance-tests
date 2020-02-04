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

	Context("deploying a Zend application with local dependencies", func() {
		It("deploying a Zend app with locally-vendored dependencies", func() {
			if cutlass.Cached {
				app = cutlass.New(filepath.Join(testdata, "zend_local_deps"))
				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(ContainSubstring("Zend Framework 2"))
			}
		})

		It("uses a proxy during staging", func() {
			AssertUsesProxyDuringStagingIfPresent(filepath.Join(testdata, "zend_local_deps"))
		})
	})

	It("deploying a Zend app with remote dependencies", func() {
		if !cutlass.Cached {
			app = cutlass.New(filepath.Join(testdata, "zend_remote_deps"))
			app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Zend Framework 2"))
		}
	})
})
