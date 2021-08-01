package config

type opts struct {
	Debug         bool
	Verbose       bool
	DBPath        string
	IsRemoteSvr   bool
	IsSynchSvr    bool
	IsLocalWebSvr bool
}

var Opts opts
