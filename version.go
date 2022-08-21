package arubaos

import "bitbucket.org/HelgeOlav/utils/version"

const Version = "0.0.1"

// init registers this module
func init() {
	m := version.ModuleVersion{
		Name:    "arubaos",
		Version: Version,
	}
	version.AddModule(m)
}
