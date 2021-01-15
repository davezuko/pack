package bundler

import (
	esbuild "github.com/evanw/esbuild/pkg/api"
)

func Bundle(file string) esbuild.BuildResult {
	defines := map[string]string{
		"process.env.NODE_ENV": "\"development\"",
	}
	result := esbuild.Build(esbuild.BuildOptions{
		Bundle:      true,
		EntryPoints: []string{file},
		Format:      esbuild.FormatESModule,
		LogLevel:    esbuild.LogLevelSilent,
		Define:      defines,
	})
	return result
}

func Transform(source []byte) ([]byte, error) {
	defines := map[string]string{
		"process.env.NODE_ENV": "\"development\"",
	}
	result := esbuild.Transform(string(source), esbuild.TransformOptions{
		Format:   esbuild.FormatESModule,
		LogLevel: esbuild.LogLevelSilent,
		Define:   defines,
	})

	// rewrite imports
	return result.Code, nil
}
