#! /bin/bash
cd src/exec/$1
# The -ldflags string is for injecting compile time and git hash data into the binary
env CC=arm-linux-gnueabihf-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags "-X main.compileDate=`date -u +%Y%m%d.%H%M%S` -X main.gitHash=`git rev-parse --verify HEAD`"
if [ ! -d "../../../armbin" ]; then
  mkdir ../../../armbin
fi
mv $1 ../../../armbin
if [ $? -eq 0 ]; then
  echo "script name okay"
  echo "success"
else
  echo "script name failed, trying lowercase"
  argOne=$(echo $1 | tr [:upper:] [:lower:])
  mv $argOne ../../../armbin/$1
  echo "lowercase success"
fi
