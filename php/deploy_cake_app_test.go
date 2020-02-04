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

	Context("deploying a Cake application with local dependencies", func() {
		It("", func() {
			if cutlass.Cached {
				app = cutlass.New(filepath.Join(testdata, "cake_local_deps"))
				app.StartCommand = "$HOME/bin/cake migrations migrate && procmgr /home/vcap/deps/org.cloudfoundry.php-web/php-web/procs.yml"
				PushAppAndConfirm(app)

				body, err := app.GetBody("/")
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(ContainSubstring("CakePHP"))
				Expect(body).ToNot(ContainSubstring("Missing Database Table"))

				Expect(app.GetBody("/users/add")).To(ContainSubstring("Add User"))
			}
		})

		It("uses a proxy during staging", func() {
			AssertUsesProxyDuringStagingIfPresent(filepath.Join(testdata, "cake_local_deps"))
		})
	})

	Context("deploying a Cake application with remote dependencies", func() {
		It("", func() {
			if !cutlass.Cached {
				app = cutlass.New(filepath.Join(testdata, "cake_remote_deps"))
				app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
				app.StartCommand = "$HOME/bin/cake migrations migrate && procmgr /home/vcap/deps/org.cloudfoundry.php-web/php-web/procs.yml"
				PushAppAndConfirm(app)

				body, err := app.GetBody("/")
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(ContainSubstring("CakePHP"))
				Expect(body).ToNot(ContainSubstring("Missing Database Table"))
				Expect(app.GetBody("/users/add")).To(ContainSubstring("Add User"))
			}
		})
	})
})
