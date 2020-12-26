@REM @echo off
:loop
cls

go build .
@REM IF %ERRORLEVEL% EQU 0 start /B /wait csgo-reporter.exe
csgo-reporter.exe

pause
goto loop
