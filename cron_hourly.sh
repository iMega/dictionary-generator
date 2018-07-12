#!/bin/bash
PATH=/usr/local/bin:/usr/local/sbin:~/bin:/usr/bin:/bin:/usr/sbin:/sbin

LINE=$(head -$((${RANDOM} % `wc -l < ~/eng/eng_words.txt` + 1)) ~/eng/eng_words_new.txt | tail -1)

TITLE="$(echo $LINE | cut -d'=' -f1)"
SUBTITLE="$(echo $LINE | cut -d'=' -f2)"
DESC="$(echo $LINE | cut -d'=' -f3)"

osascript -e "display notification \"$DESC\" with title \"$TITLE\" subtitle \"$SUBTITLE\""
