FROM robotodd-base

VOLUME ["/root/go/src/github.com/luckybroman5/http-log-reconstructor"]

WORKDIR /root/go/src/github.com/luckybroman5/http-log-reconstructor

CMD [ "go", "run", "main.go" ]