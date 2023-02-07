@REM @echo off
:loop
cls

gocritic check -enableAll -disable="#experimental,#opinionated,#commentedOutCode" ./...

::go build .
go build -v -ldflags="-w -s"

pause
goto loop
