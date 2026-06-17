package main

import "runtime/debug"

func version() string {
	v := "dev"
	if info, ok := debug.ReadBuildInfo(); ok {
		v = `v` + info.Main.Version
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				v = v + "-" + setting.Value
				break
			}
		}
	}
	return v
}
