# Homebrew Custom Tap Setup

## Creating the Custom Tap Repository

To complete the Homebrew setup, you need to create a separate repository for your tap:

### 1. Create the Tap Repository

```bash
# Create a new repository named 'homebrew-tap' on GitHub
# Repository URL: https://github.com/reinkes/homebrew-tap
```

### 2. Initialize the Tap Repository

```bash
# Clone the new repository
git clone https://github.com/reinkes/homebrew-tap.git
cd homebrew-tap

# Create the formula directory
mkdir -p Formula

# Copy the formula file
cp ../go-fancy-login/homebrew-formula.rb Formula/fancy-login-go.rb

# Commit and push
git add .
git commit -m "Add fancy-login-go formula"
git push origin main
```

### 3. Set up HOMEBREW_TOKEN

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Generate a new token with these permissions:
   - `public_repo` (for public repositories)
   - `workflow` (to trigger workflows)
3. Add the token to your repository secrets:
   - Go to your go-fancy-login repository
   - Settings → Secrets and variables → Actions
   - Add new secret: `HOMEBREW_TOKEN` with your token value

### 4. How It Works

When you create a release (tag starting with `v`):

1. **Build** - CI builds binaries for all platforms
2. **Release** - Creates GitHub release with assets
3. **Homebrew** - Automatically updates the formula in your tap with:
   - New version number
   - Updated SHA256 checksum
   - New download URL

### 5. Testing the Tap

Once set up, users can install with:

```bash
# Add the tap
brew tap reinkes/tap

# Install
brew install reinkes/tap/fancy-login-go

# Or in one command
brew install reinkes/tap/fancy-login-go
```

### 6. Formula Structure

The formula (`Formula/fancy-login-go.rb`) contains:

```ruby
class FancyLoginGo < Formula
  desc "Fancy AWS/Kubernetes login tool"
  homepage "https://github.com/reinkes/go-fancy-login"
  url "https://github.com/reinkes/go-fancy-login/archive/vX.Y.Z.tar.gz"
  sha256 "auto-generated-checksum"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args, "./cmd"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/fancy-login-go --version")
  end
end
```

### 7. Alternative: Use GitHub Releases

If you prefer, you can also configure the formula to download from GitHub releases instead of source:

```ruby
url "https://github.com/reinkes/go-fancy-login/releases/download/v#{version}/fancy-login-go-darwin-amd64.tar.gz"
```

This would be faster to install but requires more complex formula logic for different architectures.

## Troubleshooting

- **Token issues**: Ensure HOMEBREW_TOKEN has correct permissions
- **Formula errors**: Test locally with `brew install --build-from-source Formula/fancy-login-go.rb`
- **Version mismatches**: CI automatically updates version and checksum on release

## Next Steps

1. Create the `homebrew-tap` repository
2. Set up the HOMEBREW_TOKEN secret
3. Create your first release to test the automation
4. The formula will be automatically updated on future releases