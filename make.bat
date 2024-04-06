@echo off

:: ��������ļ�����
set output_name=ass
set CGO_ENABLED=0

:: ����������
set goos=linux windows
set goarch=386 amd64 arm arm64

setlocal enabledelayedexpansion

:: ��ʼ�������
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
                        echo ���ڱ��� !output_file!...
                        go build -v -o !output_file!
                  )
            ) else (
                  echo ���ڱ��� !output_file!...
                  go build -v -o !output_file!
            ) 
      )
)
endlocal

echo �������
exit /b 0
