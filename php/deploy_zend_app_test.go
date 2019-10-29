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

	// broken pending fix in php-composer-cnb to handle local vendor directories
	PIt("deploying a Zend app with locally-vendored dependencies", func() {
		app = cutlass.New(Fixtures("zend_local_deps"))
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Zend Framework 2"))
	})

	AssertNoInternetTraffic("zend_local_deps")

	// broken pending fix in php-composer-cnb where composer dependency name was wrong in buildpack.toml
	//   composer 1.8.5 seems to be broken
	PIt("deploying a Zend app with remote dependencies", func() {
		app = cutlass.New(Fixtures("zend_remote_deps"))
		app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Zend Framework 2"))
	})
})
