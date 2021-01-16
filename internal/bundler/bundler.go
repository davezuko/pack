package bundler

import (
	"fmt"
	"path"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

var rewritePackageImports = esbuild.Plugin{
	Name: "rewrite-imports",
	Setup: func(build esbuild.PluginBuild) {
		build.OnResolve(esbuild.OnResolveOptions{
			Filter:    "^[@A-z]",
			Namespace: "file",
		}, func(args esbuild.OnResolveArgs) (esbuild.OnResolveResult, error) {
			return esbuild.OnResolveResult{
				Path:      "/web_modules/" + args.Path + ".js",
				Namespace: "module",
				External:  true,
			}, nil
		})
	},
}

var markAllImportsAsExternal = esbuild.Plugin{
	Name: "mark-all-imports-as-external",
	Setup: func(build esbuild.PluginBuild) {
		build.OnResolve(esbuild.OnResolveOptions{
			Filter:    "^.",
			Namespace: "file",
		}, func(args esbuild.OnResolveArgs) (esbuild.OnResolveResult, error) {
			if path.Ext(args.Path) == "" {
				args.Path += ".js"
			}
			fmt.Printf("resolve: %s\n", args.Path)
			return esbuild.OnResolveResult{
				Path:     args.Path,
				External: true,
			}, nil
		})
	},
}

func Bundle(file string) esbuild.BuildResult {
	defines := map[string]string{
		"process.env.NODE_ENV": "\"development\"",
	}
	return esbuild.Build(esbuild.BuildOptions{
		EntryPoints: []string{file},
		Bundle:      true,
		Format:      esbuild.FormatESModule,
		LogLevel:    esbuild.LogLevelSilent,
		Define:      defines,
	})
}

func Transform(file string) esbuild.BuildResult {
	defines := map[string]string{
		"process.env.NODE_ENV": "\"development\"",
	}
	return esbuild.Build(esbuild.BuildOptions{
		EntryPoints: []string{file},
		Bundle:      true,
		Format:      esbuild.FormatESModule,
		LogLevel:    esbuild.LogLevelSilent,
		Plugins:     []esbuild.Plugin{rewritePackageImports, markAllImportsAsExternal},
		Define:      defines,
	})
}

type PackageBundler struct {
	cache map[string]esbuild.BuildResult
}

func (p *PackageBundler) Build(pkg string) esbuild.BuildResult {
	fmt.Printf("Build web module: %s\n", pkg)
	defines := map[string]string{
		"process.env.NODE_ENV": "\"development\"",
	}
	return esbuild.Build(esbuild.BuildOptions{
		Bundle:      true,
		EntryPoints: []string{pkg},
		Format:      esbuild.FormatESModule,
		LogLevel:    esbuild.LogLevelSilent,
		Define:      defines,
		Plugins:     []esbuild.Plugin{rewritePackageImports},
	})
}

func NewPackageBundler() PackageBundler {
	return PackageBundler{
		cache: make(map[string]esbuild.BuildResult),
	}
}
