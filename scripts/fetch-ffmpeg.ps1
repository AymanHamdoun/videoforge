# Downloads static ffmpeg.exe + ffprobe.exe into resources\bin\windows\ so they
# can be bundled by the NSIS installer. Run once before `wails build --nsis` (and in CI).
#
#   powershell -ExecutionPolicy Bypass -File scripts\fetch-ffmpeg.ps1
#
# Source: gyan.dev "release essentials" static build.
$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$dest = Join-Path $root "resources\bin\windows"
New-Item -ItemType Directory -Force -Path $dest | Out-Null

$tmp = New-Item -ItemType Directory -Force -Path (Join-Path $env:TEMP ("ff_" + [guid]::NewGuid()))
$zip = Join-Path $tmp "ffmpeg.zip"
$url = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"

Write-Host "-> downloading Windows ffmpeg + ffprobe (gyan.dev)..."
Invoke-WebRequest -Uri $url -OutFile $zip
Expand-Archive -Path $zip -DestinationPath $tmp -Force

$ffmpeg  = Get-ChildItem -Path $tmp -Recurse -Filter "ffmpeg.exe"  | Select-Object -First 1
$ffprobe = Get-ChildItem -Path $tmp -Recurse -Filter "ffprobe.exe" | Select-Object -First 1
Copy-Item $ffmpeg.FullName  (Join-Path $dest "ffmpeg.exe")  -Force
Copy-Item $ffprobe.FullName (Join-Path $dest "ffprobe.exe") -Force

Remove-Item -Recurse -Force $tmp
Write-Host "OK ffmpeg + ffprobe in $dest"
& (Join-Path $dest "ffmpeg.exe") -version | Select-Object -First 1
