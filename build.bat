@echo off
echo Installing dependencies...
go get github.com/wailsapp/wails/v2@latest
go get github.com/zmb3/spotify/v2@latest  
go get golang.org/x/oauth2@latest
go mod tidy

echo Building SpotLy Overlay...
wails build -clean

if %ERRORLEVEL% EQU 0 (
    if exist "build\bin\spotly.exe" copy "build\bin\spotly.exe" "spotly.exe" >nul
    echo Build successful! Run spotly.exe to start the app.
    echo Don't forget to configure your Spotify credentials in C:\Users\%USERNAME%\.spotly\config.json
) else (
    echo Build failed. Check error messages above.
    echo You may need to install Wails CLI: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    echo Then try: wails dev
)
pause
