package config

type opts struct {
	Debug         bool
	Verbose       bool
	DBPath        string
	IsRemoteSvr   bool
	IsSynchSvr    bool
	IsLocalWebSvr bool
	CreateUser    string // value here is the username
	Email         string
	Password      string
}

var Opts opts
