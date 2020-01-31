package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	Context("deploying a simple webapp written in F#", func() {
		BeforeEach(func() {
			app = cutlass.New(Fixtures("fsharp_msbuild"))
			app.Memory = "2G"
		})

		PIt("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World from F#!"))
		})
	})
})
