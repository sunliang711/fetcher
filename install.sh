#!/bin/bash
link="https://gitee.com/sunliang711/fetcher/attach_files/602033/download/fetcher-Linux-x86_64.tar.bz2"

install(){
    dest=${1:?'missing install location'}
    if [ ! -d ${dest} ];then
        echo "Create ${dest}..."
        mkdir -p ${dest}
    fi
    dest="$(realpath $dest)"

    tarFile="${link##*/}"
    dirName="${tarFile%.tar.bz2}"

    cd /tmp
    if [ ! -e ${tarFile} ];then
        echo "Download fetcher: ${link} ..."
        curl -LO "${link}" || { echo "Download failed"; exit 1; }
    fi

    tar -C ${dest} -jxvf ${tarFile}

    cd ${dest} && mv ${dirName} fetcher

}

install $1
