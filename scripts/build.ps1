# Cross-compile and package nextmail-api for multiple platforms.
#
# Produces, under dist/:
#   nextmail-api_<version>_<os>_<arch>.tar.gz   (linux / darwin)
#   nextmail-api_<version>_<os>_<arch>.zip      (windows)
#   checksums.txt                               SHA-256 of every archive
#
# Parameters:
#   -Version   build version label (default: git describe, or "dev")
#   -DistDir   output directory (default: dist)
[CmdletBinding()]
param(
    [string]$Version,
    [string]$DistDir = "dist"
)

$ErrorActionPreference = "Stop"

$Module = "nextmail-api"
$Cmd    = "./cmd/server"

# Target platforms, expressed as GOOS/GOARCH. Edit this list to suit.
$Targets = @(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Run a native command without letting its stderr trip Stop. Native tools
# (git, go, tar) write diagnostics to stderr; under Stop that would throw a
# NativeCommandError. We relax to Continue, merge stderr into the output
# stream, and let callers judge success via $LASTEXITCODE.
function Invoke-NativeCommand {
    param([scriptblock]$Block)
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        & $Block 2>&1
    } finally {
        $ErrorActionPreference = $prev
    }
}

# --- version ----------------------------------------------------------------
if (-not $Version) {
    $gitVer = (Invoke-NativeCommand { git describe --tags --always --dirty } | Out-String).Trim()
    if ($LASTEXITCODE -eq 0 -and $gitVer) { $Version = $gitVer } else { $Version = "dev" }
}

# ldflags: strip symbols + DWARF (-s -w) for smaller binaries, and stamp the
# version into internal/version.Version (a no-op until that package exists).
$LdFlags = "-s -w -X $($Module)/internal/version.Version=$Version"

# --- preflight --------------------------------------------------------------
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "go toolchain not found in PATH"
}

# Use the Windows builtin bsdtar explicitly. Resolving "tar" via PATH can pick
# up MSYS/Git GNU tar, which chokes on drive-letter paths (E:\... read as host).
$TarExe = Join-Path $env:SystemRoot "System32\tar.exe"
if (-not (Test-Path $TarExe)) {
    throw "tar.exe not found at $TarExe (needed for .tar.gz; Windows 10 1803+ ships it). Use scripts/build.sh in Git Bash/WSL as an alternative."
}

# Reset output dir.
if (Test-Path $DistDir) { Remove-Item -Recurse -Force $DistDir }
New-Item -ItemType Directory -Path $DistDir | Out-Null
$DistDir = (Resolve-Path $DistDir).Path

$checksums = Join-Path $DistDir "checksums.txt"

# Save build env so we can restore it afterwards.
$saved = @{ GOOS = $env:GOOS; GOARCH = $env:GOARCH; CGO_ENABLED = $env:CGO_ENABLED }
$env:CGO_ENABLED = "0"

try {
    foreach ($target in $Targets) {
        $os, $arch = $target -split "/"
        $bin = if ($os -eq "windows") { "$Module.exe" } else { $Module }
        $ext = if ($os -eq "windows") { "zip" } else { "tar.gz" }
        $archiveName = "${Module}_${Version}_${os}_${arch}"

        Write-Host "==> Building $os/$arch"

        $staging = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
        New-Item -ItemType Directory -Path $staging | Out-Null

        $env:GOOS = $os
        $env:GOARCH = $arch

        $out = Join-Path $staging $bin
        $buildOut = Invoke-NativeCommand { go build -trimpath -ldflags "$LdFlags" -o "$out" "$Cmd" }
        if ($LASTEXITCODE -ne 0) {
            $buildOut | Out-String | Write-Host
            throw "go build failed for $os/$arch"
        }

        if (Test-Path README.md) { Copy-Item README.md $staging }

        $archive = Join-Path $DistDir "${archiveName}.${ext}"
        if ($os -eq "windows") {
            Compress-Archive -Path "$staging\*" -DestinationPath $archive
        } else {
            Invoke-NativeCommand { & $TarExe -C "$staging" -czf "$archive" . } | Out-Null
            if ($LASTEXITCODE -ne 0) { throw "tar failed for $os/$arch" }
        }

        Remove-Item -Recurse -Force $staging

        $hash = (Get-FileHash -Algorithm SHA256 $archive).Hash
        # AppendAllText writes LF line endings (no CRLF), so the checksum file
        # verifies cleanly with `sha256sum -c` on Linux too.
        [System.IO.File]::AppendAllText($checksums, "$hash  $(Split-Path $archive -Leaf)`n")

        Write-Host "    -> $(Split-Path $archive -Leaf)"
    }
} finally {
    $env:GOOS = $saved.GOOS
    $env:GOARCH = $saved.GOARCH
    $env:CGO_ENABLED = $saved.CGO_ENABLED
}

Write-Host ""
Write-Host "Built $($Targets.Count) targets. Artifacts in ${DistDir}\:"
Get-ChildItem $DistDir | ForEach-Object { Write-Host "  $($_.Name)" }
