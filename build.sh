#!/bin/bash

file_path="$HOME/gta-api"

# make file if not exists
if [ ! -d "$file_path" ]; then
  mkdir $file_path
fi

# build go file
go build -o $file_path

# move view file
# cp -r view $file_path

cd ./batch

for file in ./*
do
  cd $file
  go build -o $file_path
  cd ..
done