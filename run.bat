@echo off

SET filename=csgo-reporter

:loop
cls

gocritic check -enableAll -disable="#experimental,#opinionated,#commentedOutCode" ./...
go build -tags debug -o %filename%.exe

IF %ERRORLEVEL% EQU 0 %filename%.exe

pause
goto loop