{
  description = "Nix Development Flake Template";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-24.11";
    nixpkgs-unstable.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix.url = "github:numtide/treefmt-nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
      flake-utils,
      treefmt-nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        name = "example.com";
        version = "0.0.1";
        pkgs = nixpkgs.legacyPackages.${system};
        unstable = nixpkgs-unstable.legacyPackages.${system};
        treefmtEval = treefmt-nix.lib.evalModule pkgs ./nix/treefmt.nix;
      in
      {
        devShells.default = import ./nix/shell.nix {
          inherit pkgs;
          inherit unstable;
        };
        packages = import ./nix/packages.nix {
          inherit pkgs;
          inherit name;
          inherit version;
        };
        formatter = treefmtEval.config.build.wrapper;
        checks = {
          formatting = treefmtEval.config.build.check self;
        };
      }
    );
}
