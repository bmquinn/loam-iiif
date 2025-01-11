# LoamIIIF

A terminal user interface (TUI) for browsing IIIF resources with integrated AI chat capabilities powered by AWS Bedrock Nova Lite.

## Installation

```bash
go install github.com/bmquinn/loam-iiif/cmd/loam-iiif@latest
```

Or clone the repository and build from source:

```bash
git clone https://github.com/bmquinn/loam-iiif.git
cd loam-iiif
go build
```

## Prerequisites

Before running LoamIIIF, ensure you have the following:

1. Go 1.21 or higher installed
2. AWS CLI v2 installed and configured
3. Active AWS SSO login session (`aws sso login`)
4. AWS account with access to Amazon Bedrock and Nova Lite service
   - Your AWS account must have Bedrock service enabled
   - Access to Amazon Nova Lite (https://aws.amazon.com/ai/generative-ai/nova/) must be granted
   - Appropriate IAM permissions to invoke Bedrock models

## AWS Setup

1. Configure AWS SSO:

   ```bash
   aws configure sso
   ```

   Follow the prompts to set up your SSO credentials.

2. Login to AWS SSO:

   ```bash
   aws sso login
   ```

   This step is required before running LoamIIIF to ensure you have valid AWS credentials.

3. Verify Bedrock Access:
   - Ensure your AWS account has Bedrock service enabled
   - Confirm you have access to Amazon Nova Lite in the Bedrock console
   - Check that you have the necessary IAM permissions:
     ```json
     {
       "Version": "2012-10-17",
       "Statement": [
         {
           "Effect": "Allow",
           "Action": ["bedrock:InvokeModel", "bedrock:ListFoundationModels"],
           "Resource": "*"
         }
       ]
     }
     ```

## Usage

### Interactive Mode

1. Ensure you have an active AWS SSO session:

   ```bash
   aws sso login
   ```

2. Run LoamIIIF:
   ```bash
   loam-iiif
   ```

### Key Bindings

- `Tab`: Switch focus between URL input and results list
- `Enter`: Open detail view or navigate into collection
- `O`: Open current item's URL in browser
- `Esc`: Close detail view or go back to previous list
- `c`: Toggle chat panel
- `Ctrl+C`: Quit application

### Command-Line Usage

You can use LoamIIIF directly from the command line by providing the `--manifest` URL and a `--prompt`. Here's an example:

```bash
loam-iiif --manifest https://api.dc.library.northwestern.edu/api/v2/collections/59ec43f9-a96c-4314-9b44-9923790b371c\?as\=iiif --prompt "Can you translate these titles into English?"
```

**Expected Output:**

```
Certainly! Here are the translated titles from Arabic to English:

1. **Arabic Manuscripts from West Africa**
   - Arabic: مخطوطات عربية من غرب أفريقيا

2 ...
```

### Chat Features

The chat panel allows you to interact with AWS Bedrock Nova Lite model to ask questions about the IIIF resources you're browsing. The chat maintains context of your current navigation and can provide insights about the collections and manifests.

## Configuration

By default, LoamIIIF uses the following AWS configuration:

- Region: us-east-1
- Profile: default AWS SSO profile

To use a different AWS SSO profile, set the AWS_PROFILE environment variable before running the application:

```bash
export AWS_PROFILE="your-sso-profile-name"
loam-iiif
```

You can also set it for a single run:

```bash
AWS_PROFILE="your-sso-profile-name" loam-iiif
```

## Troubleshooting

1. **AWS SSO Session Expired**

   - Error: "Failed to initialize ChatService" or "expired credentials"
   - Solution: Run `aws sso login` to refresh your credentials

2. **No Access to Nova Lite**

   - Error: "AccessDeniedException" when sending chat messages
   - Solution: Ensure your AWS account has access to Amazon Bedrock and Nova Lite service

3. **Invalid AWS Region**
   - Error: "model is not supported in this Region"
   - Solution: Ensure you're using us-east-1 region where Nova Lite is available

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
