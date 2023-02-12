FROM golang:1.20.0-bullseye

WORKDIR /tikoperator

COPY . .

RUN rm -rf manifests

RUN go build

CMD [ "./TrainInKubes" ]