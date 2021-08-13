package config

type opts struct {
	Debug                   bool
	Verbose                 bool
	DBPath                  string
	Port                    string
	IsRemoteSvr             bool
	IsSynchSvr              bool
	IsLocalWebSvr           bool
	CreateUser              string // value here is the username
	Email                   string
	Password                string
	Title, Descr, Body, Tag string
	QG, Q, QL, QT, QD, QB   string
	QI                      int64
	Short                   bool
	Limit                   int
	Last                    bool
	Version                 bool
	Export                  string
	Import                  string
	Update                  bool
	Delete                  bool
	WhoAmI                  bool
}

var Opts opts
