package integration_test

import (
	"fmt"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("An app deployed specifying unsupported extensions and valid", func() {
	var app *cutlass.App
	BeforeEach(func() {
		app = cutlass.New(filepath.Join(testdata, "unsupported_extensions"))
		app.SetEnv("COMPOSER_GITHUB_OAUTH_TOKEN", os.Getenv("COMPOSER_GITHUB_OAUTH_TOKEN"))
		app.SetEnv("BP_DEBUG", os.Getenv("LOG_LEVEL"))
	})
	AfterEach(func() { app = DestroyApp(app) })

	It("runs", func() {
		PushAppAndConfirm(app)

		By("should not display default php startup warning messages")

		for _, extension := range []string{"hotdog", "meatball"} {
			msg := fmt.Sprintf("NOTICE: PHP message: PHP Warning:  PHP Startup: Unable to load dynamic library '%s.so'", extension)
			Expect(app.Stdout.String()).To(ContainSubstring(msg))
		}

		By("should load the module without issue")
		Expect(app.GetBody("/")).To(ContainSubstring("curl module has been loaded successfully"))
	})
})
