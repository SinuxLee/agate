@echo off
set GO111MODULE=on
set CGO_ENABLED=0
set PROJECT=svr

rem 获取 git branch
for /F %%i in ('git rev-parse --abbrev-ref HEAD') do ( set BRANCH=%%i)

rem 获取 git commit id
for /F %%i in ('git rev-parse --short HEAD') do ( set COMMIT_ID=%%i)

rem 获取时间
set DATE_TIME=%date:~0,4%-%date:~5,2%-%date:~8,2%/%time:~0,2%:%time:~3,2%:%time:~6,2%

rem 获取 golang 版本
for /F "tokens=3 delims= " %%i in ('go version') do (set GO_VERSION=%%i)

set VER_INFO=_branch:%BRANCH%_commitid:%COMMIT_ID%_date:%DATE_TIME%_goversion:%GO_VERSION%

go build -ldflags "-X main.version=%VER_INFO%" -mod=mod -o %PROJECT%.exe cmd/%PROJECT%/main.go

if %errorlevel% gtr 0 pause
