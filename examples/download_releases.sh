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


# download_product thunderbird 45.0
# ESR45=`w3m -dump https://ftp.mozilla.org/pub/thunderbird/releases/ | cut -d' ' -f3 | \grep "45.*" | \grep -v b`
# for v in $ESR45; do
#     version=${v%/}
#     download_product thunderbird $version
# done
# download_product thunderbird 52.0
# ESR52=`w3m -dump https://ftp.mozilla.org/pub/thunderbird/releases/ | cut -d' ' -f3 | \grep "52.*" | \grep -v 52.0`
# for v in $ESR52; do
#     version=${v%/}
#     download_product thunderbird $version
# done

# download_product firefox 45.0esr
# ESR45=`w3m -dump https://ftp.mozilla.org/pub/firefox/releases/ | cut -d' ' -f3 | \grep "45.*esr" | \grep -v 45.0esr`
# for v in $ESR45; do
#     version=${v%/}
#     download_product firefox $version
# done
# download_product firefox 52.0esr
# ESR52=`w3m -dump https://ftp.mozilla.org/pub/firefox/releases/ | cut -d' ' -f3 | \grep "52.*esr" | \grep -v 52.0esr`
# for v in $ESR52; do
#     version=${v%/}
#     download_product firefox $version
# done

# download_product firefox 57
PRODUCT=$1
VERSION=$2
case $VERSION in
    nightly:*)
	locale=`echo $VERSION | cut -d':' -f2`
	filename=`w3m -dump https://ftp.mozilla.org/pub/$PRODUCT/nightly/latest-mozilla-central-l10n/ | cut -d' ' -f2 | \grep ".*\.$locale\.linux-x86_64\.tar\.bz2$" | tail -n1`
	work="pub/${PRODUCT}/nightly"
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
