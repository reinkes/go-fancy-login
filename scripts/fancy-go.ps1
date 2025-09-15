# PowerShell wrapper function for fancy-login-go
# Add this to your PowerShell profile ($PROFILE)

function fancy-go {
    param(
        [Parameter(ValueFromRemainingArguments)]
        [string[]]$Arguments
    )
    
    # Run the Go binary with all arguments and capture exit code
    & fancy-login-go.exe @Arguments
    $exitCode = $LASTEXITCODE

    # Source the AWS profile script if it exists and the command succeeded
    if ($exitCode -eq 0) {
        $profileScript = "$env:TEMP\aws_profile.ps1"
        if (Test-Path $profileScript) {
            . $profileScript
            Write-Host "✅ AWS_PROFILE environment variable updated" -ForegroundColor Green
        }
    }
}

# Create an alias for convenience
Set-Alias -Name fancy -Value fancy-go

# Export both the function and alias (not needed for regular scripts)