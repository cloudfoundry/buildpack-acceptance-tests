package integration_test

import (
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var (
		app         *cutlass.App
		fixtureName string
	)

	JustBeforeEach(func() {
		app = cutlass.New(Fixtures(fixtureName))
	})

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	Context("Deploying an app with multiple projects", func() {
		BeforeEach(func() {
			fixtureName = "multiple_projects_msbuild"
		})

		PIt("compiles both apps", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello, I'm a string!"))
			Eventually(app.Stdout.String, 10*time.Second).Should(ContainSubstring("Hello from a secondary project!"))
		})
	})

	Context("Deploying a self-contained solution with multiple projects", func() {
		BeforeEach(func() {
			fixtureName = "self_contained_solution_2.2"
		})

		It("can run the app", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})
})
