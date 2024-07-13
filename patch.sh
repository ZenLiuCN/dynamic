#!/bin/sh
if  [ ! -d "$GOROOT/src/cmd/objfile" ]; then
  cp -r $GOROOT/src/cmd/internal $GOROOT/src/cmd/objfile
  echo patched go sdk
else
  rm -rf $GOROOT/src/cmd/objfile
  echo cleaned go sdk
fi