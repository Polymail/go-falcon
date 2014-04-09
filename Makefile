ifndef FALCONGOBIN
  FALCONGOBIN=go
endif

all: deps
		$(FALCONGOBIN) install

deps:
		$(FALCONGOBIN) get launchpad.net/goyaml
		$(FALCONGOBIN) get github.com/lib/pq
		$(FALCONGOBIN) get code.google.com/p/go.text/encoding
		$(FALCONGOBIN) get github.com/garyburd/redigo/redis
		$(FALCONGOBIN) get github.com/sloonz/go-iconv
		$(FALCONGOBIN) get launchpad.net/gocheck
test:
		$(FALCONGOBIN) test -v ./...
