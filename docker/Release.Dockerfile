FROM robotodd-base

WORKDIR /root/go/src/github.com/luckybroman5/http-log-reconstructor

COPY . .

RUN go get github.com/spf13/cobra && go build

CMD [ "./http-log-reconstructor" ]