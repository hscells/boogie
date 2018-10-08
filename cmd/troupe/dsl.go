package main

type ClientConfig struct {
	Jobs []JobConfig `json:"jobs"`
}

type JobConfig struct {
	SSHUsername string `json:"username"`
	SSHAddress  string `json:"address"`
	Pipeline    string `json:"pipeline"`
	Logger      string `json:"logger"`
}
