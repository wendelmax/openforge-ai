@echo off
setlocal enabledelayedexpansion

set PORT=9090
set BASE=http://127.0.0.1:%PORT%
set REPORT=smoke-report.json
set TIMESTAMP=%DATE% %TIME%

echo === OpenForge Smoke Test ===
echo Timestamp: %TIMESTAMP%
echo.

echo === 1. Build ===
set CGO_ENABLED=0
go build -o openforge.exe ..\cmd\openforge
if %ERRORLEVEL% neq 0 (
    echo BUILD FAILED
    exit /b 1
)
for %%F in (openforge.exe) do set BUILD_SIZE=%%~zF
echo Build OK (%BUILD_SIZE% bytes)

echo === 2. Start Server ===
start "" openforge.exe serve --port %PORT%
timeout /t 3 /nobreak >nul

echo === 3. Run Tests ===
echo.

set RESULTS=[

rem ---- health ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_health.json %BASE%/health > %TEMP%\_health.meta 2>&1
call :parse "health" %TEMP%\_health.meta %TEMP%\_health.json

rem ---- v1/health ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_v1health.json %BASE%/v1/health > %TEMP%\_v1health.meta 2>&1
call :parse "v1/health" %TEMP%\_v1health.meta %TEMP%\_v1health.json

rem ---- chat ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_chat.json -X POST %BASE%/v1/chat -H "Content-Type: application/json" -d "{\"model\":\"phi-3-mini\",\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}]}" > %TEMP%\_chat.meta 2>&1
call :parse "chat" %TEMP%\_chat.meta %TEMP%\_chat.json

rem ---- completion ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_completion.json -X POST %BASE%/v1/completion -H "Content-Type: application/json" -d "{\"model\":\"phi-3-mini\",\"prompt\":\"hello world\"}" > %TEMP%\_completion.meta 2>&1
call :parse "completion" %TEMP%\_completion.meta %TEMP%\_completion.json

rem ---- embeddings ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_embed.json -X POST %BASE%/v1/embeddings -H "Content-Type: application/json" -d "{\"model\":\"bge-small\",\"input\":[\"hello world\"]}" > %TEMP%\_embed.meta 2>&1
call :parse "embeddings" %TEMP%\_embed.meta %TEMP%\_embed.json

rem ---- rerank ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_rerank.json -X POST %BASE%/v1/rerank -H "Content-Type: application/json" -d "{\"model\":\"bge-small\",\"query\":\"ai\",\"documents\":[\"doc1\",\"doc2\"]}" > %TEMP%\_rerank.meta 2>&1
call :parse "rerank" %TEMP%\_rerank.meta %TEMP%\_rerank.json

rem ---- models ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_models.json %BASE%/v1/models > %TEMP%\_models.meta 2>&1
call :parse "models" %TEMP%\_models.meta %TEMP%\_models.json

rem ---- devices ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_devices.json %BASE%/v1/devices > %TEMP%\_devices.meta 2>&1
call :parse "devices" %TEMP%\_devices.meta %TEMP%\_devices.json

rem ---- benchmark ----
curl -s -w "\nHTTP_CODE:%%{http_code}\nTIME_TOTAL:%%{time_total}\nSIZE:%%{size_download}" -o %TEMP%\_bench.json -X POST %BASE%/v1/benchmark -H "Content-Type: application/json" -d "{\"model\":\"phi-3-mini\",\"iterations\":3}" > %TEMP%\_bench.meta 2>&1
call :parse "benchmark" %TEMP%\_bench.meta %TEMP%\_bench.json

rem close JSON array
set RESULTS=%RESULTS%{"name":"__done__"}]

echo %RESULTS% > %REPORT%

echo === 4. Stop Server ===
taskkill /f /im openforge.exe >nul 2>&1

echo.
echo === Report: %REPORT% ===
echo.
type %REPORT%
echo.
echo Done.
exit /b 0

:parse
set name=%1
set metafile=%2
set jsonfile=%3

set HTTP_CODE=0
set TIME_TOTAL=0
set SIZE=0

for /f "tokens=2 delims=:" %%a in ('findstr "HTTP_CODE:" %metafile%') do set HTTP_CODE=%%a
for /f "tokens=2 delims=:" %%b in ('findstr "TIME_TOTAL:" %metafile%') do set TIME_TOTAL=%%b
for /f "tokens=2 delims=:" %%c in ('findstr "SIZE:" %metafile%') do set SIZE=%%c

set HTTP_CODE=!HTTP_CODE: =!
set TIME_TOTAL=!TIME_TOTAL: =!
set SIZE=!SIZE: =!

set RESULTS=%RESULTS%{"name":"%name%","http_code":"!HTTP_CODE!","time_sec":"!TIME_TOTAL!","size_bytes":"!SIZE!"},

echo [%name%] HTTP=!HTTP_CODE! time=!TIME_TOTAL!s size=!SIZE!B

if not "!HTTP_CODE!"=="200" (
    if not "!HTTP_CODE!"=="201" (
        echo [%name%] WARNING: unexpected HTTP !HTTP_CODE!
    )
)

exit /b 0
