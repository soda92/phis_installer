#!/bin/bash
ssh admin@192.168.1.230 "\$env:HTTP_PROXY='http://192.168.6.117:7897'; \$env:HTTPS_PROXY='http://192.168.6.117:7897'; curl -L -o nsis-3.10-setup.exe https://sourceforge.net/projects/nsis/files/NSIS%203/3.10/nsis-3.10-setup.exe/download && Start-Process -FilePath '.
sis-3.10-setup.exe' -ArgumentList '/S', '/D=C:\NSIS' -Wait"
