if (-not (Test-Path -Path .\packages)) {
    New-Item -ItemType Directory -Path .\packages
}

python38.exe -m pip download -r .\requirements.txt -d .\packages\  -i https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple
