package integration_test

import (
	"os"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF PHP Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	// Tested and working
	It("deploying a symfony 2.1 app with locally-vendored dependencies", func() {
		SkipUnlessCached()

		app = cutlass.New(Fixtures("symfony_2_local_deps"))
		PushAppAndConfirm(app)

		By("dynamically generates the content for the root route")
		Expect(app.GetBody("/")).To(ContainSubstring("Running on Symfony!"))

		By("supports Symphony app routing")
		Expect(app.GetBody("/hello/foo")).To(ContainSubstring("Hello foo!\n\nRunning on Symfony!"))
	})

	// broken picks PHP 7.3 because of composer.json PHP requirement, but the app is too old to support PHP 7.3
	PIt("deploying a symfony 2.1 app with remotely-sourced dependencies", func() {
		SkipUnlessUncached()

		app = cutlass.New(Fixtures("symfony_2_remote_deps"))
		app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
		PushAppAndConfirm(app)

		By("dynamically generates the content for the root route")
		Expect(app.GetBody("/")).To(ContainSubstring("Running on Symfony!"))
	})

	// broken pending fix in php-composer-cnb where composer dependency name was wrong in buildpack.toml
	//   composer 1.8.5 seems to be broken
	It("deploying a symfony 2.8 app", func() {
		app = cutlass.New(Fixtures("symfony_28_remote_deps"))
		app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
		PushAppAndConfirm(app)

		By("that root route has content that is dynamically generated")
		Expect(app.GetBody("/")).To(ContainSubstring("Running on Symfony!"))
	})
})
