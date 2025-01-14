{ pkgs, unstable }:
with pkgs;
mkShell {
  name = "go shell";

  buildInputs = with pkgs; [
    go
    gotools
    go-tools
    gopls
    go-outline
    gopls
    gopkgs
    gocode-gomod
    godef
    golint
  ];

  shellHook = ''
    # echo "The shell!"
    git config core.hooksPath .hooks
  '';
}
