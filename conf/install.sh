#!/usr/bin/bash

# Totem location by default is wherever the user dragged the /Totem folder, but the user can also specify a particular location.
dir=$(pwd)
echo $dir

if [ $# -eq 1 ]
    then dir="$1"
fi

export TOTEM_DIR=$dir

# work in progress...

