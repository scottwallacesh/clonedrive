# Prepare an Alpine image for Go and rclone
FROM alpine as base
WORKDIR /root
RUN apk update
RUN apk add lsof go git fuse fuse-dev alpine-sdk
RUN go get -u -v github.com/ncw/rclone
RUN mkdir bin && cp go/bin/rclone bin/

# Use the prepared base image to deploy the app
FROM base
WORKDIR /root
RUN mkdir -p mnt/cache mnt/GoogleDriveCrypt mnt/union go/src/clonedrive
COPY config .config/rclone/rclone.conf
COPY clonedrive.go go/src/clonedrive
COPY lib go/src/clonedrive/lib
COPY mounter go/src/clonedrive/mounter
COPY rclone go/src/clonedrive/rcloneWORKDIR /root/go/src/clonedrive
RUN go build clonedrive
CMD ["./clonedrive"]