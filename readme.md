# Cortado
Very simple S3 file editor that temporarily downloads a file from S3, opens it in `vim` and re-uploads it when you exit `vim`

# Usage:
```
cortado [options] {bucket}

Options:
  -editor string
        Which editor to use, only a cli editor will function properly (default "vim")
  -prefix string
        Key prefix to use
  -region string
        AWS Region (default "eu-west-1")
```
pressing `tab` will auto complete object keys. currently there is no protection from opening large or binary files.

# Installation
getting started with Cortado is easy:
```
brew tap nodefortytwo/homebrew-tap
```
Note: `brew` will need an access token for reasons... you can create one here if you have't already https://github.com/settings/tokens

then
```
brew install cortado
```

# Caveats
I used this project to brush up on golang, i'm sure it can be improved and pull requests are appreciated.