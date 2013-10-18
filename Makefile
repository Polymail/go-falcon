ifndef FALCONGOBIN
  FALCONGOBIN=go
endif

all: deps
		$(FALCONGOBIN) install

deps:
		$(FALCONGOBIN) get launchpad.net/goyaml
		$(FALCONGOBIN) get github.com/bmizerany/pq
		$(FALCONGOBIN) get code.google.com/p/mahonia
		$(FALCONGOBIN) get github.com/garyburd/redigo/redis
		$(FALCONGOBIN) get github.com/sloonz/go-qprintable
		$(FALCONGOBIN) get github.com/sloonz/go-iconv
		$(FALCONGOBIN) get launchpad.net/gocheck
test:
		$(FALCONGOBIN) test -v ./...
