@echo off
REM Start Mediasoup SFU service
setlocal enabledelayedexpansion

cd /d "C:\Users\Admin\Desktop\VTP\mediasoup-sfu"

echo Starting Mediasoup SFU service...
echo Installing dependencies...
call "C:\Program Files\nodejs\npm.cmd" install

echo.
echo Starting service on port 3000...
echo.
call "C:\Program Files\nodejs\npm.cmd" start

pause
