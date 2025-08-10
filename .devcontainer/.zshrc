export PS1='%F{blue}%~ %(?.%F{green}.%F{red})%#%f '

if [[ -e ~/.config/dotfiles/zsh/zshrc ]]; then
  source ~/.config/dotfiles/zsh/zshrc
fi

# pnpm
export PNPM_HOME="/root/.local/share/pnpm"
case ":$PATH:" in
  *":$PNPM_HOME:"*) ;;
  *) export PATH="$PNPM_HOME:$PATH" ;;
esac
# pnpm end
