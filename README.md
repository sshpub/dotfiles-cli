# dotfiles-cli

CLI tool for the [sshpub/dotfiles](https://github.com/sshpub/dotfiles) modular dotfiles framework.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/sshpub/dotfiles/main/install.sh | bash
```

## Build from source

```bash
git clone https://github.com/sshpub/dotfiles-cli.git
cd dotfiles-cli
make build
./dist/dotfiles version
```

## Usage

```bash
dotfiles setup          # First-run wizard
dotfiles module list    # List available modules
dotfiles doctor         # Health check
dotfiles version        # Print version
```

## License

Apache 2.0
