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

	BeforeEach(func() {
		SkipUnlessStack("cflinuxfs3")
		app = cutlass.New(Fixtures("nancy_kestrel_msbuild_dotnet2"))
	})

	PIt("displays a simple text homepage", func() {
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Hello from Nancy running on CoreCLR"))
	})
})
