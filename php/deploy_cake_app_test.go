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

	// broken picks PHP 7.3 because of composer.json PHP requirement, yet the fixture asks for mcrypt extension which is gone in 7.3
	//   fix is to remove mcrypt extension from .bp-config/options.json
	Context("deploying a Cake application with local dependencies", func() {
		PIt("", func() {
			SkipUnlessCached()
			app = cutlass.New(Fixtures("cake_local_deps"))
			app.StartCommand = "$HOME/bin/cake migrations migrate && $HOME/.bp/bin/start"
			PushAppAndConfirm(app)

			body, err := app.GetBody("/")
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(ContainSubstring("CakePHP"))
			Expect(body).ToNot(ContainSubstring("Missing Database Table"))

			Expect(app.GetBody("/users/add")).To(ContainSubstring("Add User"))
		})

		AssertNoInternetTraffic("cake_local_deps")
	})

	// broken picks PHP 7.3 because of composer.json PHP requirement, yet the fixture asks for mcrypt extension which is gone in 7.3
	//  fix is to remove mcrypt extension from .bp-config/options.json
	Context("deploying a Cake application with remote dependencies", func() {
		PIt("", func() {
			SkipUnlessUncached()
			app = cutlass.New(Fixtures("cake_remote_deps"))
			app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
			app.StartCommand = "$HOME/bin/cake migrations migrate && $HOME/.bp/bin/start"
			PushAppAndConfirm(app)

			body, err := app.GetBody("/")
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(ContainSubstring("CakePHP"))
			Expect(body).ToNot(ContainSubstring("Missing Database Table"))
			Expect(app.GetBody("/users/add")).To(ContainSubstring("Add User"))
		})
	})
})
