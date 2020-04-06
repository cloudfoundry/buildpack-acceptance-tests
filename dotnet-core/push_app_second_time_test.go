package integration_test

import (
	"regexp"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing an app a second time", func() {
	const (
		DownloadRegexp = `Downloading from .*\/dotnet-sdk..*\.tar\.xz`
		ReuseRegexp    = `Reusing cached download from previous build`
	)

	var app *cutlass.App

	BeforeEach(func() {
		SkipUnlessUncached()

		app = cutlass.New(Fixtures("source_2.1_float_runtime"))
		app.Buildpacks = []string{"dotnet_core_buildpack"}
	})

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		//app = DestroyApp(app)
	})

	It("uses the cache for manifest dependencies", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.ANSIStrippedString()).To(MatchRegexp(DownloadRegexp))
		Expect(app.Stdout.ANSIStrippedString()).ToNot(MatchRegexp(ReuseRegexp))

		app.Stdout.Reset()
		PushAppAndConfirm(app)
		reuseRe := regexp.MustCompile(ReuseRegexp)

		// These assertion are under the assumption the following meta buildpack is used:
		// https://github.com/cloudfoundry/dotnet-core-cnb/blob/master/compat/buildpack.toml

		reuseMatches := reuseRe.FindAllString(app.Stdout.ANSIStrippedString(), -1)
		Expect(reuseMatches).To(HaveLen(3))

		Expect(app.Stdout.ANSIStrippedString()).ToNot(MatchRegexp(DownloadRegexp))
	})
})
