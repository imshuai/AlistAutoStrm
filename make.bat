@echo off

:: 定义输出文件名称
set output_name=ass
set CGO_ENABLED=0

:: 交叉编译参数
set goos=linux windows
set goarch=386 amd64 arm arm64

setlocal enabledelayedexpansion

:: 开始交叉编译
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
                        echo 正在编译 !output_file!...
                        go build -o !output_file! > nul
                  )
            ) else (
                  echo 正在编译 !output_file!...
                  go build -o !output_file! > nul
            ) 
      )
)
endlocal

echo 编译完成
exit /b 0
