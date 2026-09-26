#!/bin/sh
# windres 的「预处理器」：不用 gcc，直接输出 .rc 原文
for a; do f="$a"; done
cat "$f"
