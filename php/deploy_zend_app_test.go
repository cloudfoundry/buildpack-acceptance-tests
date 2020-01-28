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

	// test fixture use `$_ENV` to get VCAP_SERVICES, `$_ENV` is not used anymore
	//   as it is not recomended, should use `getenv(..)` instead.
	PIt("deploying a Zend app with locally-vendored dependencies", func() {
		if cutlass.Cached {
			app = cutlass.New(filepath.Join(testdata, "zend_local_deps"))
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Zend Framework 2"))
			AssertNoInternetTraffic("zend_local_deps")
		}
	})

	// test fixture use `$_ENV` to get VCAP_SERVICES, `$_ENV` is not used anymore
	//	//   as it is not recomended, should use `getenv(..)` instead.
	PIt("deploying a Zend app with remote dependencies", func() {
		if !cutlass.Cached {
			app = cutlass.New(filepath.Join(testdata, "zend_remote_deps"))
			app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Zend Framework 2"))
		}
	})
})
