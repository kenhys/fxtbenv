#!/bin/zsh

source $(dirname $0)/fxtbenv

function fxtbenv_eval() {
    line=$1
    command=`echo "$line" | cut -d' ' -f2`
    profile=`echo "$line" | cut -d' ' -f3`
    case $command in
	use)
	    eval $line
	    ;;
	*)
	    echo "Failed to set $line"
	    :
	    ;;
    esac
}

function fxtbenv_chpwd() {
    RCFILE=$(pwd)/.fxtbenvrc
    if [ -f "$RCFILE" ]; then
	cat $RCFILE | while read LINE || [ -n "${LINE}" ]; do
	    case $LINE in
		fxtbenv*|fxenv*|tbenv*)
		    fxtbenv_eval $LINE
		    ;;
		*)
		    echo "LINE: ${LINE}"
		    ;;
	    esac
	done
    fi
}

autoload -Uz add-zsh-hook
add-zsh-hook -d precmd fxtbenv_chpwd
add-zsh-hook chpwd fxtbenv_chpwd

