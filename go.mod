module simple-raft

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Xynnn007/simple-raft v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/Xynnn007/simple-raft => ./
