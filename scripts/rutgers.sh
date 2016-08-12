#!/bin/bash

format=protobuf #protobuf or json
path=/home/tevin/go/bin/

home=/var/lib/uct/scrapers
sudo mkdir -p ${home}
campus() {
    code=$1
    OLD=${home}/rutgers-${code}-old.${format}
    LATEST=${home}/rutgers-${code}-latest.${format}
    LOG=/var/log/uct/scrapers/rutgers-${code}.log
    if [ ! -f ${LATEST} ]; then
        sudo ${path}rutgers -c ${code} -f ${format} > >(tee ${LATEST}) 2>>${LOG} | ${path}db -f ${format} >> ${LOG} 2>&1
        exit 0
    fi
    sudo cp ${LATEST} ${OLD}
    sudo ${path}rutgers -c ${code} -f ${format} > >(tee ${LATEST}) 2>>${LOG} | ${path}db -f ${format} -d ${OLD} >> ${LOG} 2>&1
}

campus CM &
campus NK &
campus NB &

wait $(jobs -p)
