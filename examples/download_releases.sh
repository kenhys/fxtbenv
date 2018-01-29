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
VERSION=57
LIST=`w3m -dump https://ftp.mozilla.org/pub/firefox/releases/ | cut -d' ' -f3 | \grep "$VERSION.*" | \grep -v esr | \grep -v 0b`
for v in $LIST; do
    version=${v%/}
    download_product firefox $version
done

