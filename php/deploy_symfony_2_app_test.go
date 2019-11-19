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

	// Tested and working
	It("deploying a symfony 2.1 app with locally-vendored dependencies", func() {
		if cutlass.Cached {
			app = cutlass.New(filepath.Join(testdata, "symfony_2_local_deps"))
			PushAppAndConfirm(app)

			By("dynamically generates the content for the root route")
			Expect(app.GetBody("/")).To(ContainSubstring("Running on Symfony!"))

			By("supports Symphony app routing")
			Expect(app.GetBody("/hello/foo")).To(ContainSubstring("Hello foo!\n\nRunning on Symfony!"))

			AssertNoInternetTraffic("symfony_2_local_deps")
		}
	})

	// broken picks PHP 7.3 because of composer.json PHP requirement, but the app is too old to support PHP 7.3
	PIt("deploying a symfony 2.1 app with remotely-sourced dependencies", func() {
		if !cutlass.Cached {
			app = cutlass.New(filepath.Join(testdata, "symfony_2_remote_deps"))
			app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
			PushAppAndConfirm(app)

			By("dynamically generates the content for the root route")
			Expect(app.GetBody("/")).To(ContainSubstring("Running on Symfony!"))
		}
	})

	// working, so long as php-composer-cnb has correct buildpack.toml, see v0.0.44 commits.
	It("deploying a symfony 2.8 app", func() {
		if !cutlass.Cached {
			app = cutlass.New(filepath.Join(testdata, "symfony_28_remote_deps"))
			app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
			app.Disk = "512M"
			PushAppAndConfirm(app)

			By("that root route has content that is dynamically generated")
			Expect(app.GetBody("/")).To(ContainSubstring("Running on Symfony!"))
		}
	})
})
