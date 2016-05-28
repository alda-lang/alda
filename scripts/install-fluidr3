#!/bin/sh

# this script fetches the latest version of FluidR3 and installs it on your machine

set -e

FLUID_VERSION=0.1.1
FLUID_NAME=fluid-r3-${FLUID_VERSION}
FLUID_WORK_FOLDER=fluidr3

echo "Downloading ${FLUID_NAME}"
curl -O https://repo1.maven.org/maven2/org/bitbucket/daveyarwood/fluid-r3/${FLUID_VERSION}/${FLUID_NAME}.jar

mkdir -p ${FLUID_WORK_FOLDER}
mv ${FLUID_NAME}.jar ${FLUID_WORK_FOLDER}
cd ${FLUID_WORK_FOLDER} && unzip -qq ${FLUID_NAME}.jar
mkdir -p ~/.gervill
cp fluid-r3.sf2 ~/.gervill/soundbank-emg.sf2
echo "Installed ${FLUID_NAME}"
cd .. && rm -rf ${FLUID_WORK_FOLDER}
