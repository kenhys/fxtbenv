#!/bin/bash

#
# Sample script to download firefox/thunderbird linux(x86-64) binary files
#

function download_product() {
    product=$1
    version=$2
    BASE=https://ftp.mozilla.org
    file="${product}-${version}.tar.bz2"
    for locale in en-US ja; do
	work="pub/${product}/releases/${version}/linux-x86_64/$locale"
	path="$work/$file"
	mkdir -p $work
	if [ ! -f $path ]; then
	    url="$BASE/$path"
	    echo "download: $url"
	    (cd $work && wget -q $url)
	else
	    echo "skip download: $path"
	fi
    done
}


if [ $# -ne 2 ]; then
    echo "Usage: $0 PRODUCT VERSION"
    echo "$ $0 firefox 57"
    exit 1
fi
# download_product firefox 57
PRODUCT=$1
VERSION=$2
case $VERSION in
    nightly:*)
	locale=`echo $VERSION | cut -d':' -f2`
	filename=`w3m -dump https://ftp.mozilla.org/pub/$PRODUCT/nightly/latest-mozilla-central-l10n/ | cut -d' ' -f2 | \grep ".*\.$locale\.linux-x86_64\.tar\.bz2$" | tail -n1`
	work="pub/${PRODUCT}/nightly/latest-mozilla-central-l10n"
	path="$work/$filename"
	mkdir -p $work
	if [ ! -f $path ]; then
	    url="https://ftp.mozilla.org/pub/$PRODUCT/nightly/latest-mozilla-central-l10n/$filename"
	    echo "download: $url"
	    (cd $work && wget -q $url)
	else
	    echo "skip download: $path"
	fi
	;;
    nightly)
	;;
    *)
	LIST=`w3m -dump https://ftp.mozilla.org/pub/$PRODUCT/releases/ | cut -d' ' -f3 | \grep "$VERSION.*" | \grep -v 0b | \grep -v funnelcake`
	for v in $LIST; do
	    version=${v%/}
	    download_product $PRODUCT $version
	done
esac
