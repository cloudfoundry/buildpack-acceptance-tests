package integration_test

import (
	"fmt"

	"github.com/Masterminds/semver"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App
	var (
		latest21RuntimeVersion, previous21RuntimeVersion string
		latest21ASPNetVersion, previous21ASPNetVersion   string
		latest21SDKVersion, previous21SDKVersion         string
		latest31SDKVersion, previous31SDKVersion         string
	)

	// BeforeEach(func() {
	// 	latest21RuntimeVersion = GetLatestDepVersion("dotnet-runtime", "2.1.x")
	// 	previous21RuntimeVersion = GetLatestDepVersion("dotnet-runtime", fmt.Sprintf("<%s", latest21RuntimeVersion))
	//
	// 	latest21ASPNetVersion = GetLatestDepVersion("dotnet-aspnetcore", "2.1.x")
	// 	previous21ASPNetVersion = GetLatestDepVersion("dotnet-aspnetcore", fmt.Sprintf("<%s", latest21ASPNetVersion))
	//
	// 	latest21SDKVersion = GetLatestDepVersion("dotnet-sdk", "2.1.x")
	// 	previous21SDKVersion = GetLatestDepVersion("dotnet-sdk", fmt.Sprintf("<%s", latest21SDKVersion))
	//
	// 	latest31SDKVersion = GetLatestDepVersion("dotnet-sdk", "3.1.x")
	// 	previous31SDKVersion = GetLatestDepVersion("dotnet-sdk", fmt.Sprintf("<%s", latest31SDKVersion))
	// })

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	Context("deploying a source-based app", func() {
		Context("with dotnet-runtime 3.1", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("simple_3.1_source"))
			})

			It("displays a simple text homepage", func() {
				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
			})
		})

		Context("with dotnet sdk 2.1 in global json", func() {
			Context("when the sdk exists", func() {
				BeforeEach(func() {
					app = ReplaceFileTemplate(Fixtures("source_2.1_global_json_templated"), "global.json", "sdk_version", latest21SDKVersion)
				})

				PIt("displays a simple text homepage", func() {
					PushAppAndConfirm(app)

					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest21SDKVersion)))
					Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.1"))
				})

			})

			Context("when the sdk is missing", func() {
				var constructedVersion string

				BeforeEach(func() {
					version, err := semver.NewVersion(latest21SDKVersion)
					Expect(err).ToNot(HaveOccurred())

					baseFeatureLine := (version.Patch() / 100) * 100
					constructedVersion = fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), baseFeatureLine)
					app = ReplaceFileTemplate(Fixtures("source_2.1_global_json_templated"), "global.json", "sdk_version", constructedVersion)
				})

				PIt("Logs a warning about using default SDK", func() {
					PushAppAndConfirm(app)
					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("SDK %s in global.json is not available", constructedVersion)))
					Expect(app.Stdout.String()).To(ContainSubstring("falling back to latest version in version line"))
					Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.1"))
				})
			})
		})

		Context("with buildpack.yml and global.json files", func() {
			Context("when SDK versions don't match", func() {
				BeforeEach(func() {
					app = ReplaceFileTemplate(Fixtures("with_buildpack_yml_templated"), "global.json", "sdk_version", previous21SDKVersion)
				})

				PIt("installs the specific version from buildpack.yml instead of global.json", func() {
					app = ReplaceFileTemplate(app.Path, "buildpack.yml", "sdk_version", previous31SDKVersion)
					app.Push()

					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", previous31SDKVersion)))
				})

				PIt("installs the floated version from buildpack.yml instead of global.json", func() {
					app = ReplaceFileTemplate(app.Path, "buildpack.yml", "sdk_version", "3.1.x")
					app.Push()

					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest31SDKVersion)))
				})
			})

			Context("when SDK version from buildpack.yml is not available", func() {
				BeforeEach(func() {
					app = ReplaceFileTemplate(Fixtures("with_buildpack_yml_templated"), "buildpack.yml", "sdk_version", "2.0.0-preview7")
				})

				PIt("fails due to missing SDK", func() {
					Expect(app.Push()).ToNot(Succeed())

					Eventually(app.Stdout.String).Should(ContainSubstring("SDK 2.0.0-preview7 in buildpack.yml is not available"))
					Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet SDK: no match found for 2.0.0-preview7"))
				})
			})
		})

		Context("with node prerendering", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("source_prerender_node"))
				app.Disk = "2G"
			})

			PIt("displays a simple text homepage", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("1 + 2 = 3"))
			})
		})

		Context("when RuntimeFrameworkVersion is explicitly defined in csproj", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(Fixtures("source_2.1_explicit_runtime_templated"), "netcoreapp2.csproj", "runtime_version", previous21RuntimeVersion)
				app.Disk = "2G"
			})

			PIt("publishes and runs, using exact runtime", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", previous21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("when RuntimeFrameworkVersion is floated in csproj", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("source_2.1_float_runtime"))
				app.Disk = "2G"
			})

			PIt("publishes and runs, using latest patch runtime", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("when the app has Microsoft.AspNetCore.All version 2.1", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("source_aspnetcore_all_2.1"))
				app.Disk = "2G"
			})

			PIt("publishes and runs, using the TargetFramework for the runtime version and the latest 2.1 patch of dotnet-aspnetcore", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("when the app has Microsoft.AspNetCore.App version 2.1", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("source_aspnetcore_app_2.1"))

				app.Disk = "2G"
			})

			PIt("publishes and runs, installing the correct runtime and aspnetcore versions", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))

				By("accepts SIGTERM and exits gracefully")
				Expect(app.Stop()).To(Succeed())
				Eventually(func() string { return app.Stdout.String() }, 30*time.Second, 1*time.Second).Should(ContainSubstring("Goodbye, cruel world!"))
			})
		})

		Context("when the app has Microsoft.AspNetCore.All version 2.0", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("source_2.0"))

				app.Disk = "1G"
			})

			PIt("publishes and runs, installing the a roll forward runtime and aspnetcore versions", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("with AssemblyName specified", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("with_dot_in_name"))
				app.Disk = "2G"
			})

			PIt("successfully pushes an app with an AssemblyName", func() {
				PushAppAndConfirm(app)
			})
		})

		Context("with libgdiplus", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("uses_libgdiplus_with_3.1"))
			})

			PIt("displays a simple text homepage", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Installing libgdiplus"))
			})
		})

		Context("without libgdiplus", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("source_aspnetcore_app_2.1"))
			})

			PIt("displays a simple text homepage", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).NotTo(ContainSubstring("Installing libgdiplus"))
			})
		})
	})

	Context("deploying an FDD app", func() {
		Context("with Microsoft.AspNetCore.App 2.1", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("fdd_aspnetcore_2.1"))

				app.Disk = "2G"
			})

			PIt("publishes and runs, and floats the runtime and aspnetcore versions by default", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))

				By("accepts SIGTERM and exits gracefully")
				Expect(app.Stop()).ToNot(HaveOccurred())
				Eventually(func() string { return app.Stdout.String() }, 30*time.Second, 1*time.Second).Should(ContainSubstring("Goodbye, cruel world!"))
			})
		})

		Context("with Microsoft.AspNetCore.App 3.0", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("fdd_3.0"))

				app.Disk = "2G"
			})

			PIt("publishes and runs, the 3.0 versions of the runtime and aspnetcore", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", "3.0")))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", "3.0")))
			})
		})

		Context("with Microsoft.AspNetCore.App 2.1 and applyPatches false", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(Fixtures("fdd_apply_patches_false_2.1_templated"), "dotnet.runtimeconfig.json", "framework_version", previous21ASPNetVersion)
			})

			PIt("installs the exact version of dotnet-aspnetcore from the runtimeconfig.json", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", previous21ASPNetVersion)))
			})
		})

		Context("with libgdiplus", func() {
			BeforeEach(func() {
				app = cutlass.New(Fixtures("uses_libgdiplus_with_3.1", "bin", "Release", "netcoreapp3.1", "linux-x64", "publish"))
			})

			PIt("displays a simple text homepage", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Installing libgdiplus"))
			})
		})
	})

	Context("deploying a self contained msbuild app with RuntimeIdentfier", func() {
		BeforeEach(func() {
			app = cutlass.New(Fixtures("self_contained_msbuild"))
		})

		PIt("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(MatchRegexp("Removing dotnet-sdk"))

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})

	Context("deploying an app with comments in the runtimeconfig.json", func() {
		It("should deploy", func() {
			app = cutlass.New(Fixtures("runtimeconfig_with_comments"))
			PushAppAndConfirm(app)
		})
	})
})
