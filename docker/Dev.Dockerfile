FROM robotodd-base

VOLUME ["/root/go/src/github.com/luckybroman5/http-log-reconstructor"]

WORKDIR /root/go/src/github.com/luckybroman5/http-log-reconstructor

RUN go get github.com/spf13/cobra

CMD [ "go", "run", "main.go" ]