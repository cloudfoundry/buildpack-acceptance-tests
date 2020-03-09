package integration_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Python Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("an app that uses miniconda and python 3", func() {
		var fixtureDir string
		BeforeEach(func() {
			var err error
			fixtureDir, err = cutlass.CopyFixture(filepath.Join(testdata, "miniconda_python_3"))
			Expect(err).ToNot(HaveOccurred())
			app = cutlass.New(fixtureDir)
			app.Disk = "2G"
			app.Memory = "1G"
			app.Buildpacks = []string{"python_buildpack"}
		})
		AfterEach(func() { _ = os.RemoveAll(fixtureDir) })

		PIt("keeps track of environment.yml", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(ContainSubstring("numpy"))

			body, err := app.GetBody("/")
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(ContainSubstring("numpy: 1.15.2"))
			Expect(body).To(ContainSubstring("python-version3"))
		})

		PIt("doesn't re-download unchanged dependencies", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("numpy"))

			app.Stdout.Reset()

			PushAppAndConfirm(app)
			// Check that numpy was not re-installed in the logs
			Expect(app.Stdout.String()).ToNot(ContainSubstring("numpy"))
		})

		PIt("it updates dependencies if environment.yml changes", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("numpy: 1.15.2"))
			Expect(app.GetBody("/")).ToNot(ContainSubstring("numpy: 1.15.0"))

			input, err := ioutil.ReadFile(filepath.Join(fixtureDir, "environment.yml"))
			Expect(err).ToNot(HaveOccurred())
			output := strings.Replace(string(input), "numpy=1.15.2", "numpy=1.15.0", 1)
			Expect(ioutil.WriteFile(filepath.Join(fixtureDir, "environment.yml"), []byte(output), 0644)).To(Succeed())

			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("numpy: 1.15.0"))
		})

		PIt("deploys offline", func() {
			AssertUsesProxyDuringStagingIfPresent("miniconda_python_3")
		})
	})
})
