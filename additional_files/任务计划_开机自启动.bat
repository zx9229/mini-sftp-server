@echo off
IF NOT EXIST "%~dp0\mini-sftp-server.exe" (
    echo please put the script in the same directory of the program.
) else (
    SET EXEC_PATH=%~dp0\mini-sftp-server.exe
    SET CONF_PATH=.\config.json
    SCHTASKS /Create /TN MINI_SFTP_SERVER /RU SYSTEM /SC ONSTART /TR "%EXEC_PATH% -conf %CONF_PATH% -offset"
    @REM 创建一个任务计划, 名为[MINI_SFTP_SERVER], 以[SYSTEM]用户运行, 在系统启动时运行,
    @REM 运行命令是"%EXEC_PATH% -conf %CONF_PATH% -offset"
)
@echo on
pause
