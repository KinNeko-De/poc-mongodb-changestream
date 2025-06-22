#!/bin/bash

# Check if protoc.exe exists
if [ ! -f ../../tools/protoc/win64/bin/protoc.exe ]; then
    echo "Error: tools/protoc/win64/bin/protoc.exe not found. Execute install-protoc-win64.sh or the corresponding script for your operating system.."
    exit 1
fi

export PATH="$PATH:$(go env GOPATH)/bin"

protobase=proto
outputpath=golang

cd ..
cd ..

mkdir -p ${outputpath}
cd ${outputpath}
outdated=( $(\ls -d */) )
for o in "${outdated[@]}"
do
    find ${o} -name *.pb.go \
    -exec rm {} +
    echo "Removed generated go code for $o"
    rm ${o}/go.mod
    rm ${o}/go.sum
    echo "Removed generated go mod files for $o"
done
cd - > /dev/null

cd ${protobase}
projects=( $(\ls -d */ ) )
cd - > /dev/null

for i in "${projects[@]}"
do
    find ${protobase}/${i} -name *.proto \
    -exec tools/protoc/win64/bin/protoc.exe --proto_path=${protobase} --go_out=${outputpath} --go_opt=paths=source_relative {} +
    echo "Generated go code for: $i"
	
	cd ${outputpath}/${i}
	modulename=$(echo github.com/kinneko-de/poc-mongodb-changestream/golang/${i} | sed 's/.$//')
	go mod init ${modulename}
	go mod tidy
	cd - > /dev/null
	echo "Generated mod files for: $i"
done

read -p "Press [Enter] to exit."
