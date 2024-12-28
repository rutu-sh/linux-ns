function create_alpine_dep {
    local dirpath=$1
    mkdir -p $dirpath
    local arch=$(uname -m)

    wget -O $dirpath/rootfs.tar.gz http://dl-cdn.alpinelinux.org/alpine/v3.21/releases/$arch/alpine-minirootfs-3.21.0-$arch.tar.gz
    cd $dirpath
    tar -xvf rootfs.tar.gz
    rm rootfs.tar.gz

    # set permissions
    chmod 755 $dirpath
    find $dirpath -type d -exec chmod 755 {} \;
}


"$@"