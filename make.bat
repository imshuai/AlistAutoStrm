@echo off

:: ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½Ä¼ï¿½ï¿½ï¿½ï¿½ï¿?
set output_name=ass
set CGO_ENABLED=0

:: ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½
set goos=linux windows
set goarch=386 amd64 arm arm64

setlocal enabledelayedexpansion

:: ï¿½ï¿½Ê¼ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿?
for %%o in (%goos%) do (
      for %%a in (%goarch%) do (
            set output_file=bin\%output_name%_%%o_%%a
            set GOOS=%%o
            set GOARCH=%%a
            if %%a==arm (
                  set GOARM=7
                  set output_file=!output_file!v7
            )
            if %%o==windows (
                  if %%a==amd64 (
                        set output_file=!output_file!.exe
                        echo ÕıÔÚ±àÒë !output_file!...
                        go build -o !output_file! > nul
                  )
            ) else (
                  echo ÕıÔÚ±àÒë !output_file!...
                  go build -o !output_file! > nul
            ) 
      )
)
endlocal

echo ½»²æ±àÒëÍê³É
exit /b 0
