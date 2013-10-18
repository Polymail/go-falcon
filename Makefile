ifndef GOBIN
  GOBIN=go
endif

all: deps
		$(GOBIN) install

deps:
		$(GOBIN) get launchpad.net/goyaml
		$(GOBIN) get github.com/bmizerany/pq
		$(GOBIN) get code.google.com/p/mahonia
		$(GOBIN) get github.com/garyburd/redigo/redis
		$(GOBIN) get github.com/sloonz/go-qprintable
		$(GOBIN) get github.com/sloonz/go-iconv
		$(GOBIN) get launchpad.net/gocheck
test:
		$(GOBIN) test -v ./...
