![MPL-2.0](https://img.shields.io/badge/License-MPL2.0-green.svg?style=flat)
[![Build Status](https://travis-ci.org/kenhys/fxtbenv.svg?branch=master)](https://travis-ci.org/kenhys/fxtbenv)

# Fxtbenv

Firefox/Thunderbird environment manager.

## Why fxtbenv?

Need to switch multiple version of Firefox/Thunderbird for verifying difference of functinality among them.

## Requirements

* go 1.8.7 or later

## How to Install fxtbenv

```
$ git clone https://github.com/kenhys/fxtbenv $HOME/.fxtbenv
$ cd $HOME/.fxtbenv
$ make
$ source $HOME/.fxtbenv/scripts/fxtbenv.zsh
```

## Usage

First, you need to install Firefox/Thunderbird. If you want to localized version, specify `VERSION:LOCALE`.

To install Firefox 57.0.4 with Japanese edition, execute the following command.

```
% fxenv install 57.0.4:ja
```

Then, create profile for it. To use `test` profile, execute the following command.
Note that `-c` option  must be specified only at first time because there is no profile yet.

```
% fxenv use 57.0.4:ja@test -c
```

Okay, now you are ready for it. Just launch Firefox.

```
% firefox
```

## License

MPL-2.0
