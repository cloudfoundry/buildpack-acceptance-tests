package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("override yml", func() {
	var (
		app           *cutlass.App
		buildpackName string
	)

	AfterEach(func() {
		if buildpackName != "" {
			cutlass.DeleteBuildpack(buildpackName)
		}
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	BeforeEach(func() {
		if !ApiHasMultiBuildpack() {
			Skip("Multi buildpack support is required")
		}

		buildpackName = "override_yml_" + cutlass.RandStringRunes(5)
		Expect(cutlass.CreateOrUpdateBuildpack(buildpackName, Fixtures("overrideyml_bp"), "")).To(Succeed())

		app = cutlass.New(Fixtures("console_app"))
		app.Buildpacks = []string{buildpackName + "_buildpack", "dotnet_core_buildpack"}
	})

	PIt("Forces dotnet-sdk from override buildpack", func() {
		Expect(app.Push()).ToNot(Succeed())
		Eventually(app.Stdout.String).Should(ContainSubstring("-----> OverrideYML Buildpack"))
		Expect(app.ConfirmBuildpack(packagedBuildpack.Version)).To(Succeed())

		Eventually(app.Stdout.String).Should(ContainSubstring("-----> Installing dotnet-sdk"))
		Eventually(app.Stdout.String).Should(MatchRegexp("Copy .*/dotnet-sdk.tgz"))
		Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet SDK: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc"))
	})
})
