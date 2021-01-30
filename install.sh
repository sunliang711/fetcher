#!/bin/bash
set -e

linuxARMLink="https://gitee.com/sunliang711/fetcher/attach_files/603298/download/fetcher-linux-arm64.tar.bz2"
linuxAMD64Link="https://gitee.com/sunliang711/fetcher/attach_files/603299/download/fetcher-linux-amd64.tar.bz2"

install(){
    dest=${1:?'missing install location'}
    echo "dest: ${dest}"
    if [ ! -d ${dest} ];then
        echo "Create ${dest}..."
        mkdir -p ${dest}
    fi
    dest="$(realpath $dest)"

    case $(uname) in
        Linux)
            case $(uname -m) in
                aarch64)
                    link="${linuxARMLink}"
                    ;;
                x86_64)
                    link="${linuxAMD64Link}"
                    ;;
            esac
        ;;
        *)
            echo "Only support Linux currently"
            exit 1
        ;;
    esac

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
