ifndef GOBIN
  GOBIN=go
endif

all:
		$(GOBIN) get launchpad.net/goyaml
		$(GOBIN) get github.com/bmizerany/pq
		$(GOBIN) get code.google.com/p/mahonia
		$(GOBIN) get github.com/garyburd/redigo/redis
		$(GOBIN) get github.com/sloonz/go-qprintable
		$(GOBIN) get github.com/sloonz/go-iconv
		$(GOBIN) get launchpad.net/gocheck
		$(GOBIN) install
test:
		$(GOBIN) test -v ./...
