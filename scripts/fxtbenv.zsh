#!/bin/zsh

source $(dirname $0)/fxtbenv

function fxtbenv_chpwd() {
    RCFILE=$(pwd)/.fxtbenvrc
    if [ -f "$RCFILE" ]; then
	line=`\grep "fxtbenv" $RCFILE`
	if [ $? -eq 0 ]; then
	    command=`echo $line | cut -d' ' -f2`
	    profile=`echo $line | cut -d' ' -f3`
	    case $command in
		use)
		    eval $line
		    ;;
		*)
		    :
		    ;;
	    esac
	else
	fi
    fi
}

autoload -Uz add-zsh-hook
add-zsh-hook -d precmd fxtbenv_chpwd
add-zsh-hook chpwd fxtbenv_chpwd

