#/bin/bash
set -e
go build
./cder cd \
  --repo https://github.com/untillpro/directcd-test \
  -t 10 \
  -w .tmp \
  -o directcd-test.exe
  -- a1 a2 a3