@echo off
setlocal EnableDelayedExpansion

REM DBSage One-Click Installation Script (Windows)
REM Database AI Assistant - One-Click Installation Script for Windows
REM 
REM Supported Windows versions: Windows 10, Windows 11, Windows Server 2019+
REM Supported architectures: amd64, arm64
REM
REM Usage:
REM   1. Download this script
REM   2. Right-click "Run as administrator" or run in PowerShell/CMD
REM   3. Follow the prompts to complete installation

REM Configuration variables
set "REPO_URL=https://github.com/murongg/dbsage"
set "BINARY_NAME=dbsage.exe"
set "INSTALL_DIR=%ProgramFiles%\DBSage"
set "CONFIG_DIR=%USERPROFILE%\.dbsage"
set "VERSION=latest"
set "FORCE_INSTALL=false"
set "INSTALL_GLOBAL=true"

REM Color codes
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "PURPLE=[95m"
set "CYAN=[96m"
set "NC=[0m"

REM Icons (using ASCII characters)
set "SUCCESS=[+]"
set "ERROR=[X]"
set "INFO=[i]"
set "WARNING=[!]"
set "ROCKET=[*]"

echo.
echo %PURPLE%===============================================================%NC%
echo %PURPLE%                  DBSage Database AI Assistant%NC%
echo %PURPLE%                     One-Click Installation%NC%
echo %PURPLE%===============================================================%NC%
echo.

REM Parse command line arguments
:parse_args
if "%~1"=="" goto start_install
if "%~1"=="-h" goto show_help
if "%~1"=="--help" goto show_help
if "%~1"=="-f" (
    set "FORCE_INSTALL=true"
    shift
    goto parse_args
)
if "%~1"=="--force" (
    set "FORCE_INSTALL=true"
    shift
    goto parse_args
)
if "%~1"=="--local" (
    set "INSTALL_GLOBAL=false"
    set "INSTALL_DIR=%CD%"
    shift
    goto parse_args
)
if "%~1"=="-v" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_args
)
if "%~1"=="--version" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_args
)
echo %RED%%ERROR% Unknown option: %~1%NC%
goto show_help

:show_help
echo DBSage One-Click Installation Script (Windows)
echo.
echo Usage: %~nx0 [options]
echo.
echo Options:
echo   -h, --help          Show help information
echo   -v, --version VER   Specify version (default: latest)
echo   -f, --force         Force reinstallation
echo   --local             Install to current directory
echo.
echo Examples:
echo   %~nx0                # Standard installation
echo   %~nx0 --local        # Local installation
echo   %~nx0 -f             # Force reinstallation
echo.
pause
exit /b 0

:start_install
REM Check administrator privileges
call :check_admin
if !errorlevel! neq 0 goto end

REM Check existing installation
call :check_existing
if !errorlevel! neq 0 goto end

REM Check system requirements
call :check_requirements
if !errorlevel! neq 0 goto end

REM Get target version
if "%VERSION%"=="latest" (
    call :get_latest_version
    if !errorlevel! neq 0 goto end
    set "TARGET_VERSION=!LATEST_VERSION!"
    echo %INFO% Target version: !TARGET_VERSION!
) else (
    set "TARGET_VERSION=%VERSION%"
    echo %INFO% Target version: !TARGET_VERSION!
)

REM Create temporary directory
call :create_temp_dir
if !errorlevel! neq 0 goto end

REM Download binary
call :download_binary
if !errorlevel! neq 0 goto cleanup

REM Install binary
call :install_binary
if !errorlevel! neq 0 goto cleanup

REM Create configuration files
call :create_config

REM Setup environment variables
call :setup_environment

REM Verify installation
call :verify_installation
if !errorlevel! neq 0 goto cleanup

REM Show post-installation instructions
call :show_post_install

goto cleanup

:check_admin
echo %INFO% Checking administrator privileges...
net session >nul 2>&1
if !errorlevel! equ 0 (
    echo %GREEN%%SUCCESS% Administrator privileges check passed%NC%
    exit /b 0
) else (
    if "%INSTALL_GLOBAL%"=="true" (
        echo %RED%%ERROR% Administrator privileges required for global installation%NC%
        echo %YELLOW%%WARNING% Please right-click "Run as administrator" or use --local option%NC%
        exit /b 1
    ) else (
        echo %GREEN%%SUCCESS% Local installation mode, no administrator privileges required%NC%
        exit /b 0
    )
)

:check_existing
if "%FORCE_INSTALL%"=="true" (
    exit /b 0
)

where dbsage >nul 2>&1
if !errorlevel! equ 0 (
    echo %YELLOW%%WARNING% DBSage is already installed on the system%NC%
    dbsage --version 2>nul || echo Current version: Unknown version
    echo.
    echo Use --force option to reinstall
    exit /b 1
)
exit /b 0

:check_requirements
echo %INFO% Checking system requirements...

REM Check PowerShell (for downloading files)
where powershell >nul 2>&1
if !errorlevel! neq 0 (
    echo %RED%%ERROR% PowerShell not found%NC%
    echo PowerShell is required for downloading files
    exit /b 1
)
echo %GREEN%%SUCCESS% PowerShell available%NC%

REM Check tar (Windows 10 build 17063 and later)
where tar >nul 2>&1
if !errorlevel! equ 0 (
    echo %GREEN%%SUCCESS% tar available%NC%
    set "HAS_TAR=true"
) else (
    echo %YELLOW%%WARNING% tar not available, will use PowerShell extraction%NC%
    set "HAS_TAR=false"
)

echo %GREEN%%SUCCESS% All system requirements check passed%NC%
exit /b 0

:get_latest_version
echo %INFO% Getting latest release version...
set "API_URL=https://api.github.com/repos/murongg/dbsage/releases/latest"

REM Use PowerShell to get latest version
for /f "usebackq delims=" %%i in (`powershell -Command "(Invoke-RestMethod -Uri '%API_URL%').tag_name" 2^>nul`) do set "LATEST_VERSION=%%i"

if "!LATEST_VERSION!"=="" (
    echo %RED%%ERROR% Failed to get latest version from GitHub API%NC%
    exit /b 1
)

echo %GREEN%%SUCCESS% Latest version: !LATEST_VERSION!%NC%
exit /b 0

:create_temp_dir
echo %INFO% Creating temporary directory...
set "TEMP_DIR=%TEMP%\dbsage_install_%RANDOM%"
mkdir "!TEMP_DIR!" 2>nul
if !errorlevel! neq 0 (
    echo %RED%%ERROR% Unable to create temporary directory%NC%
    exit /b 1
)
echo %GREEN%%SUCCESS% Temporary directory created: !TEMP_DIR!%NC%
exit /b 0

:download_binary
echo %INFO% Downloading DBSage binary...
cd /d "!TEMP_DIR!"

REM Determine platform
set "PLATFORM=windows_amd64"
echo %PROCESSOR_ARCHITECTURE% | findstr /i "arm64" >nul
if !errorlevel! equ 0 set "PLATFORM=windows_arm64"

echo %INFO% Platform: !PLATFORM!

REM Determine archive name and download URL
set "ARCHIVE_NAME=dbsage_!PLATFORM!.zip"
set "DOWNLOAD_URL=https://github.com/murongg/dbsage/releases/download/!TARGET_VERSION!/!ARCHIVE_NAME!"

echo %INFO% Download URL: !DOWNLOAD_URL!

REM Setup cache directory
set "CACHE_DIR=%USERPROFILE%\.cache\dbsage"
set "CACHED_FILE=%CACHE_DIR%\!TARGET_VERSION!_!ARCHIVE_NAME!"

if not exist "%CACHE_DIR%" mkdir "%CACHE_DIR%"

REM Check for cached version
if exist "!CACHED_FILE!" (
    echo %INFO% Using cached binary: !CACHED_FILE!
    copy "!CACHED_FILE!" "!ARCHIVE_NAME!" >nul
) else (
    REM Download using PowerShell
    echo %INFO% Downloading archive...
    powershell -Command "Invoke-WebRequest -Uri '!DOWNLOAD_URL!' -OutFile '!ARCHIVE_NAME!'" 2>nul
    if !errorlevel! neq 0 (
        echo %RED%%ERROR% Failed to download binary from !DOWNLOAD_URL!%NC%
        echo %INFO% Available releases: https://github.com/murongg/dbsage/releases%NC%
        exit /b 1
    )
    
    REM Cache the downloaded file
    copy "!ARCHIVE_NAME!" "!CACHED_FILE!" >nul
    if !errorlevel! equ 0 (
        echo %INFO% Binary cached for future installations
    )
)

REM Extract archive
echo %INFO% Extracting binary...
if "!HAS_TAR!"=="true" (
    REM Use tar if available (Windows 10+)
    tar -xf "!ARCHIVE_NAME!"
) else (
    REM Use PowerShell to extract
    powershell -Command "Expand-Archive -Path '!ARCHIVE_NAME!' -DestinationPath '.' -Force" 2>nul
)

if !errorlevel! neq 0 (
    echo %RED%%ERROR% Failed to extract archive%NC%
    exit /b 1
)

REM Verify binary exists
if not exist "%BINARY_NAME%" (
    echo %RED%%ERROR% Binary not found after extraction%NC%
    exit /b 1
)

echo %GREEN%%SUCCESS% Binary download and extraction completed%NC%
exit /b 0

:install_binary
echo %INFO% Installing DBSage...

if "%INSTALL_GLOBAL%"=="true" (
    REM Create installation directory
    if not exist "%INSTALL_DIR%" (
        mkdir "%INSTALL_DIR%"
        if !errorlevel! neq 0 (
            echo %RED%%ERROR% Unable to create installation directory %INSTALL_DIR%%NC%
            exit /b 1
        )
    )
    
    REM Copy binary file
    copy "!TEMP_DIR!\%BINARY_NAME%" "%INSTALL_DIR%\" >nul
    if !errorlevel! neq 0 (
        echo %RED%%ERROR% Unable to copy file to %INSTALL_DIR%%NC%
        exit /b 1
    )
    
    echo %GREEN%%SUCCESS% DBSage installed to %INSTALL_DIR%\%BINARY_NAME%%NC%
) else (
    copy "!TEMP_DIR!\%BINARY_NAME%" ".\%BINARY_NAME%" >nul
    if !errorlevel! neq 0 (
        echo %RED%%ERROR% Unable to copy file to current directory%NC%
        exit /b 1
    )
    
    echo %GREEN%%SUCCESS% DBSage installed to %CD%\%BINARY_NAME%%NC%
)

exit /b 0

:create_config
echo %INFO% Creating configuration directory and files...

REM Create configuration directory
if not exist "%CONFIG_DIR%" (
    mkdir "%CONFIG_DIR%"
    if !errorlevel! neq 0 (
        echo %YELLOW%%WARNING% Unable to create configuration directory %CONFIG_DIR%%NC%
        exit /b 0
    )
)

REM Create example configuration file
(
echo # DBSage Configuration File
echo # Please modify the following configuration as needed
echo.
echo # OpenAI API Configuration
echo OPENAI_API_KEY=your_openai_api_key_here
echo OPENAI_BASE_URL=https://api.openai.com/v1
echo.
echo # Database Configuration (optional, can also be added at runtime)
echo # DATABASE_URL=postgres://username:password@localhost:5432/database?sslmode=disable
echo.
echo # Log Level (optional)
echo # LOG_LEVEL=info
echo.
echo # Other Configuration
echo # MAX_CONNECTIONS=10
echo # TIMEOUT=30s
) > "%CONFIG_DIR%\config.env"

REM Create connection configuration file
echo {}> "%CONFIG_DIR%\connections.json"

echo %GREEN%%SUCCESS% Configuration files created in %CONFIG_DIR%\%NC%
echo %INFO% Please edit %CONFIG_DIR%\config.env file to set your OpenAI API Key%NC%
exit /b 0

:setup_environment
if "%INSTALL_GLOBAL%"=="true" (
    echo %INFO% Adding DBSage to system PATH...
    
    REM Check if PATH already contains installation directory
    echo %PATH% | findstr /i "%INSTALL_DIR%" >nul
    if !errorlevel! neq 0 (
        REM Add to system PATH
        setx PATH "%PATH%;%INSTALL_DIR%" /M >nul 2>&1
        if !errorlevel! equ 0 (
            echo %GREEN%%SUCCESS% PATH updated%NC%
        ) else (
            echo %YELLOW%%WARNING% Unable to automatically update PATH, please manually add %INSTALL_DIR% to system PATH%NC%
        )
    ) else (
        echo %GREEN%%SUCCESS% %INSTALL_DIR% already in PATH%NC%
    )
) else (
    echo %INFO% Local installation mode, please manually add %CD% to PATH or use full path to run%NC%
)
exit /b 0

:verify_installation
echo %INFO% Verifying installation...

if "%INSTALL_GLOBAL%"=="true" (
    "%INSTALL_DIR%\%BINARY_NAME%" --version >nul 2>&1
    if !errorlevel! equ 0 (
        echo %GREEN%%SUCCESS% Installation verification passed%NC%
        exit /b 0
    ) else (
        echo %RED%%ERROR% Installation verification failed%NC%
        exit /b 1
    )
) else (
    ".\%BINARY_NAME%" --version >nul 2>&1
    if !errorlevel! equ 0 (
        echo %GREEN%%SUCCESS% Installation verification passed%NC%
        exit /b 0
    ) else (
        echo %RED%%ERROR% Installation verification failed%NC%
        exit /b 1
    )
)

:show_post_install
echo.
echo %GREEN%%SUCCESS% 🎉 DBSage installation completed!%NC%
echo.
echo ===============================================================
echo %YELLOW%%ROCKET% Quick Start:%NC%
echo.
echo 1. %BLUE%Configure OpenAI API Key:%NC%
echo    Edit configuration file: %GREEN%%CONFIG_DIR%\config.env%NC%
echo    Set: %CYAN%OPENAI_API_KEY=your_actual_api_key%NC%
echo.
echo 2. %BLUE%Start DBSage:%NC%
if "%INSTALL_GLOBAL%"=="true" (
    echo    %GREEN%dbsage%NC%
    echo    ^(If command not found, please reopen command prompt^)
) else (
    echo    %GREEN%.\dbsage%NC%
    echo    Or add current directory to PATH and use: %GREEN%dbsage%NC%
)
echo.
echo 3. %BLUE%Add database connection:%NC%
echo    Run in DBSage: %CYAN%/add mydb%NC%
echo    Then follow the prompts to enter database connection information
echo.
echo ===============================================================
echo %YELLOW%%INFO% More Information:%NC%
echo.
echo ^• %BLUE%Configuration directory:%NC% %CONFIG_DIR%
echo ^• %BLUE%Installation directory:%NC% %INSTALL_DIR%
echo ^• %BLUE%Documentation:%NC% https://github.com/murongg/dbsage
echo ^• %BLUE%Issue reports:%NC% https://github.com/murongg/dbsage/issues
echo.
echo %GREEN%Thank you for using DBSage!%NC%
echo.
exit /b 0

:cleanup
echo %INFO% Cleaning up temporary files...
if exist "!TEMP_DIR!" (
    rd /s /q "!TEMP_DIR!" 2>nul
)

:end
echo.
pause
exit /b 0
