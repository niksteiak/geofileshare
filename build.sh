#!/bin/bash

# Arguments
#   d: Deployment directory
#

while getopts d:v: option
do
    case "${option}"
        in
        d)directory=${OPTARG};;
    esac
done

if [[ $directory = "" ]]
then
   directory="./dist"
   if [ ! -d "$directory" ]; then
       mkdir $directory
    fi
fi

source version
version=$MAJOR.$MINOR.$BUILD
echo "Directory: $directory"
echo "Version:   $version"

# Build the executable with version specified
go build -ldflags "-X main.GFSVersion=$version"

tar -zcvf $directory/gfs.tar.gz geofileshare static/ templates/
