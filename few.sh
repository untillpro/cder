#/bin/bash
set -e
go build
./cder cd \
  --repo https://github.com/untillpro/directcd-test \
  --replace https://github.com/untillpro/directcd-test-print=https://github.com/maxim-ge/directcd-test-print \
  -o out.exe \
  -t 10 \
  -w .tmp \
  -- --option1 arg1 arg2

