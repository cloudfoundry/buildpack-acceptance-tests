package integration_test

import (
	"os"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App

	BeforeEach(SkipUnlessCached)

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	Context("The app is portable", func() {
		var fixture string
		BeforeEach(func() {
			SkipUnlessStack("cflinuxfs3")
			fixture = "fdd_asp_vendored_2.1"

			app = cutlass.New(Fixtures(fixture))
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})

		It("is not connected to the internet", func() {
			AssertNoInternetTraffic(Fixtures(fixture))
		})
	})

	Context("The app is self contained", func() {
		var fixture string
		BeforeEach(func() {
			SkipUnlessStack("cflinuxfs3")
			fixture = "self_contained_2.1"
			app = cutlass.New(Fixtures(fixture))
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})

		It("is not connected to the internet", func() {
			AssertNoInternetTraffic(Fixtures(fixture))
		})
	})

	Context("The app is self contained and a preview version", func() {
		var fixture string
		BeforeEach(func() {
			if os.Getenv("CF_STACK") == "cflinuxfs2" {
				Skip("Dotnet3 only works on cflinuxfs3")
			}

			app = cutlass.New(Fixtures("self_contained_3.0"))
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Welcome"))
		})

		It("is not connected to the internet", func() {
			AssertNoInternetTraffic(Fixtures(fixture))
		})
	})
})
