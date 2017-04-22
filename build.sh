#!/bin/bash
sum="sha1sum"

if ! hash sha1sum 2>/dev/null; then
    if ! hash sha256sum 2>/dev/null; then
        echo "I can't see 'sha1sum' or 'sha256sum'"
        echo "Please install one of them!"
        exit
    fi
    sum="sha256sum"
fi

UPX=false
if hash upx 2>/dev/null; then
    UPX=true
fi

SOFTWARE="broken-pantsu"
RELEASE="./release"
SOURCE="main.go"
VERSION=`date -u +%y%m%d`
LDFLAGS="-X main.VERSION=$VERSION -s -w"
GCFLAGS=""

mkdir -p "$RELEASE"

OSES=(linux darwin windows freebsd)
ARCHS=(amd64 386)
for os in ${OSES[@]}; do
    for arch in ${ARCHS[@]}; do
        suffix=""
        if [ "$os" == "windows" ]; then
            suffix=".exe"
        fi

        env CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ${SOFTWARE}_${os}_${arch}${suffix} $SOURCE

        tar -zcf ${RELEASE}/${SOFTWARE}-${os}-${arch}-$VERSION.tar.gz ${SOFTWARE}_${os}_${arch}${suffix}
        cd release
        $sum ${SOFTWARE}-${os}-${arch}-$VERSION.tar.gz
        cd ..
        rm ${SOFTWARE}_${os}_${arch}${suffix}
    done
done


# ARM
ARMS=(5 6 7)
for v in ${ARMS[@]}; do
    env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=$v go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ${SOFTWARE}_linux_arm$v $SOURCE
done

if $UPX; then
    upx -9 ${SOFTWARE}_linux_arm*
fi
tar -zcf ${RELEASE}/${SOFTWARE}-linux-arm-$VERSION.tar.gz ${SOFTWARE}_linux_arm*
cd release
$sum ${SOFTWARE}-linux-arm-$VERSION.tar.gz
cd ..
rm ${SOFTWARE}_linux_arm*


#MIPS32LE
env CGO_ENABLED=0 GOOS=linux GOARCH=mipsle go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ${SOFTWARE}_linux_mipsle $SOURCE
env CGO_ENABLED=0 GOOS=linux GOARCH=mips go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ${SOFTWARE}_linux_mips $SOURCE

if $UPX; then
    upx -9 ${SOFTWARE}_linux_mips*
fi

tar -zcf ${RELEASE}/${SOFTWARE}-linux-mipsle-$VERSION.tar.gz ${SOFTWARE}_linux_mipsle
tar -zcf ${RELEASE}/${SOFTWARE}-linux-mips-$VERSION.tar.gz ${SOFTWARE}_linux_mips
cd release
$sum ${SOFTWARE}-linux-mipsle-$VERSION.tar.gz
$sum ${SOFTWARE}-linux-mips-$VERSION.tar.gz
cd ..
rm ${SOFTWARE}_linux_mipsle
rm ${SOFTWARE}_linux_mips
