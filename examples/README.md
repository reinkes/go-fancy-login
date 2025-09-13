# Configuration Templates

This directory contains example configuration templates for AWS and Kubernetes that work with fancy-login.

## Quick Setup

To install the templates (only if configs don't already exist):

```bash
make install-templates
```

This will:
- Copy `aws-config.template` to `~/.aws/config` (if it doesn't exist)
- Copy `kube-config.template` to `~/.kube/config` (if it doesn't exist)
- Warn you to customize the values

## Manual Setup

If you prefer to set up configurations manually:

### AWS Configuration

1. Copy the template:
   ```bash
   cp examples/aws-config.template ~/.aws/config
   ```

2. Edit `~/.aws/config` and replace:
   - `YOUR_PROJECT` with your actual project name
   - `YOUR_SSO_DOMAIN` with your AWS SSO domain
   - `YOUR_*_ACCOUNT_ID` with actual AWS account IDs
   - `YourRoleName` with actual IAM role names

### Kubernetes Configuration

1. Copy the template:
   ```bash
   cp examples/kube-config.template ~/.kube/config
   ```

2. Edit `~/.kube/config` and replace:
   - `YOUR_*_CLUSTER_*` with actual cluster information
   - `YOUR_REGION` with your AWS region
   - `YOUR_PROJECT_*_PROFILE` with AWS profile names from your AWS config

3. Get cluster information with:
   ```bash
   aws eks describe-cluster --name YOUR_CLUSTER_NAME --region YOUR_REGION
   ```

## Important Notes

- **Security**: Templates contain placeholder values only - no sensitive data
- **Profiles with "_DEV_"**: Will automatically trigger ECR login in fancy-login
- **AWS Profiles**: Must match between AWS config and Kubernetes config
- **Existing Configs**: The installer will not overwrite existing configurations

## Template Customization

Both templates are designed to work with:
- AWS SSO authentication
- EKS clusters
- Multiple environment profiles (dev, staging, prod)

Modify the templates to match your organization's:
- Naming conventions
- Account structure  
- Regional setup
- Role assignments