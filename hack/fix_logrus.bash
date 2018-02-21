#!/bin/bash
set -eo pipefail

cd $(git rev-parse --show-toplevel)
grep -rl 'Sirupsen' ./vendor/ |\
    xargs sed -i 's#Sirupsen#sirupsen#g'
